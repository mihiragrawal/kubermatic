/*
Copyright 2017-2020 by the contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authenticator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/clientauthentication"
	clientauthv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
)

// Identity is returned on successful Verify() results. It contains a parsed
// version of the AWS identity used to create the token.
type Identity struct {
	// ARN is the raw Amazon Resource Name returned by sts:GetCallerIdentity
	ARN string

	// CanonicalARN is the Amazon Resource Name converted to a more canonical
	// representation. In particular, STS assumed role ARNs like
	// "arn:aws:sts::ACCOUNTID:assumed-role/ROLENAME/SESSIONNAME" are converted
	// to their IAM ARN equivalent "arn:aws:iam::ACCOUNTID:role/NAME"
	CanonicalARN string

	// AccountID is the 12 digit AWS account number.
	AccountID string

	// UserID is the unique user/role ID (e.g., "AROAAAAAAAAAAAAAAAAAA").
	UserID string

	// SessionName is the STS session name (or "" if this is not a
	// session-based identity). For EC2 instance roles, this will be the EC2
	// instance ID (e.g., "i-0123456789abcdef0"). You should only rely on it
	// if you trust that _only_ EC2 is allowed to assume the IAM Role. If IAM
	// users or other roles are allowed to assume the role, they can provide
	// (nearly) arbitrary strings here.
	SessionName string

	// The AWS Access Key ID used to authenticate the request.  This can be used
	// in conjunction with CloudTrail to determine the identity of the individual
	// if the individual assumed an IAM role before making the request.
	AccessKeyID string
}

const (
	// The sts GetCallerIdentity request is valid for 15 minutes regardless of this parameters value after it has been
	// signed, but we set this unused parameter to 60 for legacy reasons (we check for a value between 0 and 60 on the
	// server side in 0.3.0 or earlier).  IT IS IGNORED.  If we can get STS to support x-amz-expires, then we should
	// set this parameter to the actual expiration, and make it configurable.
	requestPresignParam = 60
	// The actual token expiration (presigned STS urls are valid for 15 minutes after timestamp in x-amz-date).
	presignedURLExpiration = 15 * time.Minute
	v1Prefix               = "k8s-aws-v1."
	maxTokenLenBytes       = 1024 * 4
	clusterIDHeader        = "x-k8s-aws-id"
	// Format of the X-Amz-Date header used for expiration
	// https://golang.org/pkg/time/#pkg-constants
	dateHeaderFormat   = "20060102T150405Z"
	kindExecCredential = "ExecCredential"
	execInfoEnvKey     = "KUBERNETES_EXEC_INFO"
)

// Token is generated and used by Kubernetes client-go to authenticate with a Kubernetes cluster.
type Token struct {
	Token      string
	Expiration time.Time
}

// GetTokenOptions is passed to GetWithOptions to provide an extensible get token interface
type GetTokenOptions struct {
	Region               string
	ClusterID            string
	AssumeRoleARN        string
	AssumeRoleExternalID string
	SessionName          string
	Session              *session.Session
}

// FormatError is returned when there is a problem with token that is
// an encoded sts request.  This can include the url, data, action or anything
// else that prevents the sts call from being made.
type FormatError struct {
	message string
}

func (e FormatError) Error() string {
	return "input token was not properly formatted: " + e.message
}

// STSError is returned when there was either an error calling STS or a problem
// processing the data returned from STS.
type STSError struct {
	message string
}

func (e STSError) Error() string {
	return "sts getCallerIdentity failed: " + e.message
}

// NewSTSError creates a error of type STS.
func NewSTSError(m string) STSError {
	return STSError{message: m}
}

var parameterWhitelist = map[string]bool{
	"action":               true,
	"version":              true,
	"x-amz-algorithm":      true,
	"x-amz-credential":     true,
	"x-amz-date":           true,
	"x-amz-expires":        true,
	"x-amz-security-token": true,
	"x-amz-signature":      true,
	"x-amz-signedheaders":  true,
}

// this is the result type from the GetCallerIdentity endpoint
type getCallerIdentityWrapper struct {
	GetCallerIdentityResponse struct {
		GetCallerIdentityResult struct {
			Account string `json:"Account"`
			Arn     string `json:"Arn"`
			UserID  string `json:"UserId"`
		} `json:"GetCallerIdentityResult"`
		ResponseMetadata struct {
			RequestID string `json:"RequestId"`
		} `json:"ResponseMetadata"`
	} `json:"GetCallerIdentityResponse"`
}

// Generator provides new tokens for the AWS IAM Authenticator.
type Generator interface {
	// Get a token using credentials in the default credentials chain.
	Get(string) (Token, error)
	// GetWithRole creates a token by assuming the provided role, using the credentials in the default chain.
	GetWithRole(clusterID, roleARN string) (Token, error)
	// GetWithRoleForSession creates a token by assuming the provided role, using the provided session.
	GetWithRoleForSession(clusterID string, roleARN string, sess *session.Session) (Token, error)
	// Get a token using the provided options
	GetWithOptions(options *GetTokenOptions) (Token, error)
	// GetWithSTS returns a token valid for clusterID using the given STS client.
	GetWithSTS(clusterID string, stsAPI stsiface.STSAPI) (Token, error)
	// FormatJSON returns the client auth formatted json for the ExecCredential auth
	FormatJSON(Token) string
}

type generator struct {
	forwardSessionName bool
}

// NewGenerator creates a Generator and returns it.
func NewGenerator(forwardSessionName bool) (Generator, error) {
	return generator{
		forwardSessionName: forwardSessionName,
	}, nil
}

// Get uses the directly available AWS credentials to return a token valid for
// clusterID. It follows the default AWS credential handling behavior.
func (g generator) Get(clusterID string) (Token, error) {
	return g.GetWithOptions(&GetTokenOptions{ClusterID: clusterID})
}

// GetWithRole assumes the given AWS IAM role and returns a token valid for
// clusterID. If roleARN is empty, behaves like Get (does not assume a role).
func (g generator) GetWithRole(clusterID string, roleARN string) (Token, error) {
	return g.GetWithOptions(&GetTokenOptions{
		ClusterID:     clusterID,
		AssumeRoleARN: roleARN,
	})
}

// GetWithRoleForSession assumes the given AWS IAM role for the given session and behaves
// like GetWithRole.
func (g generator) GetWithRoleForSession(clusterID string, roleARN string, sess *session.Session) (Token, error) {
	return g.GetWithOptions(&GetTokenOptions{
		ClusterID:     clusterID,
		AssumeRoleARN: roleARN,
		Session:       sess,
	})
}

// StdinStderrTokenProvider gets MFA token from standard input.
func StdinStderrTokenProvider() (string, error) {
	var v string
	fmt.Fprint(os.Stderr, "Assume Role MFA token code: ")
	_, err := fmt.Scanln(&v)
	return v, err
}

// GetWithOptions takes a GetTokenOptions struct, builds the STS client, and wraps GetWithSTS.
// If no session has been passed in options, it will build a new session. If an
// AssumeRoleARN was passed in then assume the role for the session.
func (g generator) GetWithOptions(options *GetTokenOptions) (Token, error) {
	if options.ClusterID == "" {
		return Token{}, fmt.Errorf("ClusterID is required")
	}

	if options.Session == nil {
		// create a session with the "base" credentials available
		// (from environment variable, profile files, EC2 metadata, etc)
		sess, err := session.NewSessionWithOptions(session.Options{
			AssumeRoleTokenProvider: StdinStderrTokenProvider,
			SharedConfigState:       session.SharedConfigEnable,
		})
		if err != nil {
			return Token{}, fmt.Errorf("could not create session: %v", err)
		}
		sess.Handlers.Build.PushFrontNamed(request.NamedHandler{
			Name: "authenticatorUserAgent",
			Fn:   request.MakeAddToUserAgentHandler("aws-iam-authenticator", "v0.5.8-forked"),
		})
		if options.Region != "" {
			sess = sess.Copy(aws.NewConfig().WithRegion(options.Region).WithSTSRegionalEndpoint(endpoints.RegionalSTSEndpoint))
		}

		options.Session = sess
	}

	// use an STS client based on the direct credentials
	stsAPI := sts.New(options.Session)

	// if a roleARN was specified, replace the STS client with one that uses
	// temporary credentials from that role.
	if options.AssumeRoleARN != "" {
		var sessionSetters []func(*stscreds.AssumeRoleProvider)

		if options.AssumeRoleExternalID != "" {
			sessionSetters = append(sessionSetters, func(provider *stscreds.AssumeRoleProvider) {
				provider.ExternalID = &options.AssumeRoleExternalID
			})
		}

		if g.forwardSessionName {
			// If the current session is already a federated identity, carry through
			// this session name onto the new session to provide better debugging
			// capabilities
			resp, err := stsAPI.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if err != nil {
				return Token{}, err
			}

			userIDParts := strings.Split(*resp.UserId, ":")
			if len(userIDParts) == 2 {
				sessionSetters = append(sessionSetters, func(provider *stscreds.AssumeRoleProvider) {
					provider.RoleSessionName = userIDParts[1]
				})
			}
		} else if options.SessionName != "" {
			sessionSetters = append(sessionSetters, func(provider *stscreds.AssumeRoleProvider) {
				provider.RoleSessionName = options.SessionName
			})
		}

		// create STS-based credentials that will assume the given role
		creds := stscreds.NewCredentials(options.Session, options.AssumeRoleARN, sessionSetters...)

		// create an STS API interface that uses the assumed role's temporary credentials
		stsAPI = sts.New(options.Session, &aws.Config{Credentials: creds})
	}

	return g.GetWithSTS(options.ClusterID, stsAPI)
}

// GetWithSTS returns a token valid for clusterID using the given STS client.
func (g generator) GetWithSTS(clusterID string, stsAPI stsiface.STSAPI) (Token, error) {
	// generate an sts:GetCallerIdentity request and add our custom cluster ID header
	request, _ := stsAPI.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	request.HTTPRequest.Header.Add(clusterIDHeader, clusterID)

	// Sign the request.  The expires parameter (sets the x-amz-expires header) is
	// currently ignored by STS, and the token expires 15 minutes after the x-amz-date
	// timestamp regardless.  We set it to 60 seconds for backwards compatibility (the
	// parameter is a required argument to Presign(), and authenticators 0.3.0 and older are expecting a value between
	// 0 and 60 on the server side).
	// https://github.com/aws/aws-sdk-go/issues/2167
	presignedURLString, err := request.Presign(requestPresignParam)
	if err != nil {
		return Token{}, err
	}

	// Set token expiration to 1 minute before the presigned URL expires for some cushion
	tokenExpiration := time.Now().Local().Add(presignedURLExpiration - 1*time.Minute)
	// TODO: this may need to be a constant-time base64 encoding
	return Token{v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(presignedURLString)), tokenExpiration}, nil
}

// FormatJSON formats the json to support ExecCredential authentication
func (g generator) FormatJSON(token Token) string {
	apiVersion := clientauthv1beta1.SchemeGroupVersion.String()
	env := os.Getenv(execInfoEnvKey)
	if env != "" {
		cred := &clientauthentication.ExecCredential{}
		if err := json.Unmarshal([]byte(env), cred); err == nil {
			apiVersion = cred.APIVersion
		}
	}

	expirationTimestamp := metav1.NewTime(token.Expiration)
	execInput := &clientauthv1beta1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kindExecCredential,
		},
		Status: &clientauthv1beta1.ExecCredentialStatus{
			ExpirationTimestamp: &expirationTimestamp,
			Token:               token.Token,
		},
	}
	enc, _ := json.Marshal(execInput)
	return string(enc)
}

// Verifier validates tokens by calling STS and returning the associated identity.
type Verifier interface {
	Verify(token string) (*Identity, error)
}

type tokenVerifier struct {
	client            *http.Client
	clusterID         string
	validSTShostnames map[string]bool
}

func stsHostsForPartition(partitionID string) map[string]bool {
	validSTShostnames := map[string]bool{}

	var partition *endpoints.Partition
	for _, p := range endpoints.DefaultPartitions() {
		if partitionID == p.ID() {
			partition = &p
			break
		}
	}
	if partition == nil {
		logrus.Errorf("Partition %s not valid", partitionID)
		return validSTShostnames
	}
	stsSvc, ok := partition.Services()["sts"]
	if !ok {
		logrus.Errorf("STS service not found in partition %s", partitionID)
		return validSTShostnames
	}
	for epName, ep := range stsSvc.Endpoints() {
		rep, err := ep.ResolveEndpoint(endpoints.STSRegionalEndpointOption)
		if err != nil {
			logrus.WithError(err).Errorf("Error resolving endpoint for %s in partition %s", epName, partitionID)
			continue
		}
		parsedURL, err := url.Parse(rep.URL)
		if err != nil {
			logrus.WithError(err).Errorf("Error parsing STS URL %s", rep.URL)
			continue
		}
		validSTShostnames[parsedURL.Hostname()] = true
	}
	return validSTShostnames
}

// NewVerifier creates a Verifier that is bound to the clusterID and uses the default http client.
func NewVerifier(clusterID string, partitionID string) Verifier {
	return tokenVerifier{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		clusterID:         clusterID,
		validSTShostnames: stsHostsForPartition(partitionID),
	}
}

// verify a sts host, doc: http://docs.amazonaws.cn/en_us/general/latest/gr/rande.html#sts_region
func (v tokenVerifier) verifyHost(host string) error {
	if _, ok := v.validSTShostnames[host]; !ok {
		return FormatError{fmt.Sprintf("unexpected hostname %q in pre-signed URL", host)}
	}
	return nil
}

// Verify a token is valid for the specified clusterID. On success, returns an
// Identity that contains information about the AWS principal that created the
// token. On failure, returns nil and a non-nil error.
func (v tokenVerifier) Verify(token string) (*Identity, error) {
	if len(token) > maxTokenLenBytes {
		return nil, FormatError{"token is too large"}
	}

	if !strings.HasPrefix(token, v1Prefix) {
		return nil, FormatError{fmt.Sprintf("token is missing expected %q prefix", v1Prefix)}
	}

	// TODO: this may need to be a constant-time base64 decoding
	tokenBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, v1Prefix))
	if err != nil {
		return nil, FormatError{err.Error()}
	}

	parsedURL, err := url.Parse(string(tokenBytes))
	if err != nil {
		return nil, FormatError{err.Error()}
	}

	if parsedURL.Scheme != "https" {
		return nil, FormatError{fmt.Sprintf("unexpected scheme %q in pre-signed URL", parsedURL.Scheme)}
	}

	if err = v.verifyHost(parsedURL.Host); err != nil {
		return nil, err
	}

	if parsedURL.Path != "/" {
		return nil, FormatError{"unexpected path in pre-signed URL"}
	}

	queryParamsLower := make(url.Values)
	queryParams, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return nil, FormatError{"malformed query parameter"}
	}

	for key, values := range queryParams {
		if !parameterWhitelist[strings.ToLower(key)] {
			return nil, FormatError{fmt.Sprintf("non-whitelisted query parameter %q", key)}
		}
		if len(values) != 1 {
			return nil, FormatError{"query parameter with multiple values not supported"}
		}
		queryParamsLower.Set(strings.ToLower(key), values[0])
	}

	if queryParamsLower.Get("action") != "GetCallerIdentity" {
		return nil, FormatError{"unexpected action parameter in pre-signed URL"}
	}

	if !hasSignedClusterIDHeader(&queryParamsLower) {
		return nil, FormatError{fmt.Sprintf("client did not sign the %s header in the pre-signed URL", clusterIDHeader)}
	}

	// We validate x-amz-expires is between 0 and 15 minutes (900 seconds) although currently pre-signed STS URLs, and
	// therefore tokens, expire exactly 15 minutes after the x-amz-date header, regardless of x-amz-expires.
	expires, err := strconv.Atoi(queryParamsLower.Get("x-amz-expires"))
	if err != nil || expires < 0 || expires > 900 {
		return nil, FormatError{fmt.Sprintf("invalid X-Amz-Expires parameter in pre-signed URL: %d", expires)}
	}

	date := queryParamsLower.Get("x-amz-date")
	if date == "" {
		return nil, FormatError{"X-Amz-Date parameter must be present in pre-signed URL"}
	}

	// Obtain AWS Access Key ID from supplied credentials
	accessKeyID := strings.Split(queryParamsLower.Get("x-amz-credential"), "/")[0]

	dateParam, err := time.Parse(dateHeaderFormat, date)
	if err != nil {
		return nil, FormatError{fmt.Sprintf("error parsing X-Amz-Date parameter %s into format %s: %s", date, dateHeaderFormat, err.Error())}
	}

	now := time.Now()
	expiration := dateParam.Add(presignedURLExpiration)
	if now.After(expiration) {
		return nil, FormatError{fmt.Sprintf("X-Amz-Date parameter is expired (%.f minute expiration) %s", presignedURLExpiration.Minutes(), dateParam)}
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	req.Header.Set(clusterIDHeader, v.clusterID)
	req.Header.Set("accept", "application/json")

	response, err := v.client.Do(req)
	if err != nil {
		// special case to avoid printing the full URL if possible
		if urlErr, ok := err.(*url.Error); ok {
			return nil, NewSTSError(fmt.Sprintf("error during GET: %v", urlErr.Err))
		}
		return nil, NewSTSError(fmt.Sprintf("error during GET: %v", err))
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, NewSTSError(fmt.Sprintf("error reading HTTP result: %v", err))
	}

	if response.StatusCode != 200 {
		return nil, NewSTSError(fmt.Sprintf("error from AWS (expected 200, got %d). Body: %s", response.StatusCode, string(responseBody[:])))
	}

	var callerIdentity getCallerIdentityWrapper
	err = json.Unmarshal(responseBody, &callerIdentity)
	if err != nil {
		return nil, NewSTSError(err.Error())
	}

	// parse the response into an Identity
	id := &Identity{
		ARN:         callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.Arn,
		AccountID:   callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.Account,
		AccessKeyID: accessKeyID,
	}
	id.CanonicalARN, err = CanonicalizeARN(id.ARN)
	if err != nil {
		return nil, NewSTSError(err.Error())
	}

	// The user ID is either UserID:SessionName (for assumed roles) or just
	// UserID (for IAM User principals).
	userIDParts := strings.Split(callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.UserID, ":")
	if len(userIDParts) == 2 {
		id.UserID = userIDParts[0]
		id.SessionName = userIDParts[1]
	} else if len(userIDParts) == 1 {
		id.UserID = userIDParts[0]
	} else {
		return nil, STSError{fmt.Sprintf(
			"malformed UserID %q",
			callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.UserID)}
	}

	return id, nil
}

func hasSignedClusterIDHeader(paramsLower *url.Values) bool {
	signedHeaders := strings.Split(paramsLower.Get("x-amz-signedheaders"), ";")
	for _, hdr := range signedHeaders {
		if strings.ToLower(hdr) == strings.ToLower(clusterIDHeader) {
			return true
		}
	}
	return false
}

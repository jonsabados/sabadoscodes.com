package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/jonsabados/sabadoscodes.com/httputil"
)

const GoogleCertEndpoint = "https://www.googleapis.com/oauth2/v1/certs"

type googleAuthenticator struct {
	certsLock  sync.Mutex
	certs      GooglePublicCerts
	fetchCerts GoogleCertFetcher
	clientID   string
	getRoles   RoleOracle
}

func (a *googleAuthenticator) currentCerts(ctx context.Context) GooglePublicCerts {
	a.certsLock.Lock()
	defer a.certsLock.Unlock()

	if a.certs.Expiration.Before(time.Now()) {
		newCerts, err := a.fetchCerts(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error fetching certs")
		} else {
			a.certs = newCerts
		}
	}

	return a.certs
}

func (a *googleAuthenticator) authenticate(ctx context.Context, token string) (Principal, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("token", sanitizeTokenForLog(token)).Msg("attempting google authentication")

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Principal{}, garbageTokenError(token, "format")
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		logger.Debug().Err(err).Msg("error decoding token signature")
		return Principal{}, garbageTokenError(token, "signature malformed")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		logger.Debug().Err(err).Msg("error decoding token paylad")
		return Principal{}, garbageTokenError(token, "payload malformed")
	}

	signedContent := parts[0] + "." + parts[1]
	hash := sha256.New()
	hash.Write([]byte(signedContent))
	hashSum := hash.Sum(nil)

	valid := false
	for _, c := range a.currentCerts(ctx).Certs {
		err := rsa.VerifyPKCS1v15(c.PublicKey.(*rsa.PublicKey), crypto.SHA256, hashSum, signature)
		if err == nil {
			valid = true
			break
		} else {
			logger.Debug().Err(err).Str("cert", base64.StdEncoding.EncodeToString(c.Signature)).Msg("jwt verification failed using cert")
		}
	}
	if !valid {
		return Principal{}, errors.Errorf("invalid signature on token %s", token)
	}

	payload := new(googleToken)
	err = json.Unmarshal(payloadBytes, payload)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to unmarshal json portion of signed token")
		return Principal{}, garbageTokenError(token, "payload not json")
	}

	if payload.Aud != a.clientID {
		logger.Info().Interface("token", payload).Msg("someone passed token with invalid audience")
		return Principal{}, errors.Errorf("invalid audience: %s", payload.Aud)
	}

	if payload.Exp < time.Now().Unix() {
		return Principal{}, errors.Errorf("expired token, expiration: %s", time.Unix(payload.Exp, 0).Format(time.RFC3339))
	}

	return Principal{
		UserID: payload.Sub,
		Email:  payload.Email,
		Name:   payload.Name,
		Roles:  a.getRoles(ctx, payload.Email),
	}, nil
}

func sanitizeTokenForLog(token string) string {
	if len(token) > 8 {
		return token[0:8] + strings.Repeat("x", len(token)-8)
	}
	return token
}

func NewGoogleAuthenticator(clientID string, fetchCerts GoogleCertFetcher, roleOracle RoleOracle) Authenticator {
	authenticator := &googleAuthenticator{
		clientID:   clientID,
		fetchCerts: fetchCerts,
		getRoles:   roleOracle,
	}

	return authenticator.authenticate
}

// GooglePublicCerts carry the public certificates that can be used to validate JWT tokens signed by google, as well
// as an expiration date after which they should refresh.
type GooglePublicCerts struct {
	Certs      []*x509.Certificate
	Expiration time.Time
}

// GoogleCertFetcher is used to fetch GooglePublicCerts
type GoogleCertFetcher func(ctx context.Context) (GooglePublicCerts, error)

// NewGoogleCertFetcher creates a fully wired GoogleCertFetcher using the provided url and http client. The production
func NewGoogleCertFetcher(url string, newHttpClient httputil.HTTPClientFactory) GoogleCertFetcher {
	return func(ctx context.Context) (GooglePublicCerts, error) {
		httpClient := newHttpClient(ctx)
		res, err := httpClient.Get(url)
		if err != nil {
			return GooglePublicCerts{}, errors.WithStack(err)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return GooglePublicCerts{}, errors.WithStack(err)
		}
		if res.StatusCode != http.StatusOK {
			return GooglePublicCerts{}, CertFetchingError{res.StatusCode, string(body)}
		}
		// Thu, 25 Jun 2020 02:12:50 GMT
		expiresStr := res.Header.Get("Expires")
		expireDate, err := time.Parse(time.RFC1123, expiresStr)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("header", expiresStr).Msg("invalid expiration date header received")
			// this isn't really fatal, if we can read the rest of the response just warn and give an expiration of now
			expireDate = time.Now()
		}
		// response is json with cert key => cert pem so just read into a map since keys aren't known
		mappedCerts := make(map[string]string)
		err = json.Unmarshal(body, &mappedCerts)
		if err != nil {
			return GooglePublicCerts{}, errors.WithStack(err)
		}
		if len(mappedCerts) == 0 {
			zerolog.Ctx(ctx).Info().Str("responseBody", string(body)).Msg("no certs found")
			return GooglePublicCerts{}, errors.New("no certs found in response body")
		}
		certs := make([]*x509.Certificate, len(mappedCerts))
		i := 0
		for _, v := range mappedCerts {
			cert, err := parseCert(v)
			if err != nil {
				return GooglePublicCerts{}, errors.WithStack(err)
			}
			certs[i] = cert
			i++
		}
		return GooglePublicCerts{
			Certs:      certs,
			Expiration: expireDate,
		}, nil
	}
}

func parseCert(cert string) (*x509.Certificate, error) {
	pemVal, rest := pem.Decode([]byte(cert))
	if len(rest) > 0 {
		return nil, errors.New(fmt.Sprintf("multiple certs or invalid body in %s", cert))
	}
	parsedCert, err := x509.ParseCertificate(pemVal.Bytes)
	if err != nil {
		return nil, err
	}
	return parsedCert, nil
}

// CertFetchingError represents an error when reading googles public cert if the http status is not as expected
type CertFetchingError struct {
	StatusCode   int
	ResponseBody string
}

// Error formats the CertFetchingError into a string with status code and response body
func (c CertFetchingError) Error() string {
	return fmt.Sprintf("unexpected response code: %d, body: %s", c.StatusCode, c.ResponseBody)
}

type googleToken struct {
	Iss           string `json:"iss"`
	Azp           string `json:"azp"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	AtHash        string `json:"at_hash"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
	Iat           int64  `json:"iat"`
	Exp           int64  `json:"exp"`
	Jti           string `json:"jti"`
}

func garbageTokenError(reason string, token string) error {
	return errors.Errorf("garbage token: %s (%s)", token, reason)
}

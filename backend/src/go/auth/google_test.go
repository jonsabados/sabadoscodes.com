package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"
)

func Test_sanitizeTokenForLog(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			"long",
			"foobarblahbazweefunfoo",
			"foobarblxxxxxxxxxxxxxx",
		},
		{
			"short",
			"aba",
			"aba",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.expected, sanitizeTokenForLog(tc.input))
		})
	}
}

func Test_NewGoogleCertFetcher_NetworkError(t *testing.T) {
	asserter := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		asserter.Fail("wtf?")
	}))
	ts.Close()

	testInstance := NewGoogleCertFetcher(ts.URL, httputil.DefaultHttpClient)
	_, err := testInstance(context.Background())
	asserter.Error(err)
}

func Test_NewGoogleCertFetcher_NetworkGood(t *testing.T) {
	testCases := []struct {
		desc             string
		responseCode     int
		responseFixture  string
		expirationHeader string
		expectedSerials  []string
		expectError      bool
	}{
		{
			"internal server error",
			http.StatusInternalServerError,
			"fixture/error.json",
			"Thu, 25 Jun 2020 12:01:01 MST",
			nil,
			true,
		},
		{
			"garbage response",
			http.StatusOK,
			"fixture/garbage.html",
			"Thu, 25 Jun 2020 12:01:01 MST",
			nil,
			true,
		},
		{
			"unexpected json",
			http.StatusOK,
			"fixture/error.json",
			"Thu, 25 Jun 2020 12:01:01 MST",
			nil,
			true,
		},
		{
			"empty json",
			http.StatusOK,
			"fixture/empty.json",
			"Thu, 25 Jun 2020 12:01:01 MST",
			nil,
			true,
		},
		{
			"happy path",
			http.StatusOK,
			"fixture/googlecerts.json",
			"Thu, 25 Jun 2020 12:01:01 MST",
			[]string{
				"swjeSaDD1xCYlGMO3fPrpieIhRXTHfWqEL0EA7L3JMKPs2Dae3P/vtqN2qL7fq3Ft48xCz0swmE5Ci8OEBQZi+RB+A4t0MxMO9K3LJk1wmqyZdj0d7LZ3WFq5hyym7dQzes/4z/4UcYMel/z/jjmKM7qBvtm8a68vpAcZooMy/f13hIotTdYPwJ8fACB8EYOYVzz0gyKPFAXbXvNC64dR2IF4lR0/ql9IdgZkxqCeCyf/KQtNQ3D4p8yqvdMcJV0Va3r8Teh72zyj1U/QLnCJVURL/ircP3UDGZzN7bym/r5JQhuOHjGWTqPTsGgV0/ZkQA4pOxOvt1PUO0F1UsQTQ==",
				"uK2uXX3c28Xpjyx0rUjmC7cBSJ5j7OUJfL4EQsZbXm1I514GD+GCnn/UhYqirv3hTdH0F22aiGJdgDwofZBr5iKAVf4Z2VHaQ8sE1taMH+cAqZEquJLmDuRTRKoJh6ZW116+8cuAVtDdfBGH8INTy8hedusJh+uUTqO+xg/dEt8EQHQlvO4DlQc5iqV/dAb1TnAdl9SyKV68naxts/B+Cy8P1FrVv7LHcXBDHYTo8jquhZRnz+GuxKrhqS2W8Nyfqj+k9xYZqd/usvvu6XUmb/wDDatw9i8zUDURKulcUeCA7OKyOGjNr6pKIkKnMPDHDoCA6N6aTrZBG1fuj3G8eg==",
			},
			false,
		},
		{
			"missing expiration",
			http.StatusOK,
			"fixture/googlecerts.json",
			"",
			[]string{
				"swjeSaDD1xCYlGMO3fPrpieIhRXTHfWqEL0EA7L3JMKPs2Dae3P/vtqN2qL7fq3Ft48xCz0swmE5Ci8OEBQZi+RB+A4t0MxMO9K3LJk1wmqyZdj0d7LZ3WFq5hyym7dQzes/4z/4UcYMel/z/jjmKM7qBvtm8a68vpAcZooMy/f13hIotTdYPwJ8fACB8EYOYVzz0gyKPFAXbXvNC64dR2IF4lR0/ql9IdgZkxqCeCyf/KQtNQ3D4p8yqvdMcJV0Va3r8Teh72zyj1U/QLnCJVURL/ircP3UDGZzN7bym/r5JQhuOHjGWTqPTsGgV0/ZkQA4pOxOvt1PUO0F1UsQTQ==",
				"uK2uXX3c28Xpjyx0rUjmC7cBSJ5j7OUJfL4EQsZbXm1I514GD+GCnn/UhYqirv3hTdH0F22aiGJdgDwofZBr5iKAVf4Z2VHaQ8sE1taMH+cAqZEquJLmDuRTRKoJh6ZW116+8cuAVtDdfBGH8INTy8hedusJh+uUTqO+xg/dEt8EQHQlvO4DlQc5iqV/dAb1TnAdl9SyKV68naxts/B+Cy8P1FrVv7LHcXBDHYTo8jquhZRnz+GuxKrhqS2W8Nyfqj+k9xYZqd/usvvu6XUmb/wDDatw9i8zUDURKulcUeCA7OKyOGjNr6pKIkKnMPDHDoCA6N6aTrZBG1fuj3G8eg==",
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			asserter := assert.New(t)

			expectedPath := "/testingfun"

			ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				defer request.Body.Close()

				asserter.Equal(expectedPath, request.URL.Path)

				if tc.expirationHeader != "" {
					writer.Header().Add("Expires", tc.expirationHeader)
				}

				writer.WriteHeader(tc.responseCode)
				res, err := ioutil.ReadFile(tc.responseFixture)
				if asserter.NoError(err) {
					_, err = writer.Write(res)
					asserter.NoError(err)
				}
			}))
			defer ts.Close()

			testInstance := NewGoogleCertFetcher(fmt.Sprintf("%s%s", ts.URL, expectedPath), httputil.DefaultHttpClient)
			res, err := testInstance(context.Background())
			if tc.expectError {
				asserter.Error(err)
			} else {
				if !asserter.NoError(err) {
					return
				}
				serials := make([]string, len(res.Certs))
				for i, cert := range res.Certs {
					serials[i] = base64.StdEncoding.EncodeToString(cert.Signature)
				}
				sort.Strings(serials)
				asserter.Equal(tc.expectedSerials, serials)

				if tc.expirationHeader != "" {
					expectedTime, err := time.Parse(time.RFC1123, tc.expirationHeader)
					if asserter.NoError(err) {
						asserter.Equal(expectedTime, res.Expiration)
					}
				} else {
					asserter.WithinDuration(time.Now(), res.Expiration, time.Second*1)
				}
			}
		})
	}
}

func TestNewGoogleAuthenticator_HappyPath_FirstKey(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}
	keyTwo, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
				createCert(keyTwo, 456, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())

	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`

	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyOne, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}

	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role{
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	res, err := testInstance(context.Background(), jwt)

	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)
}

func TestNewGoogleAuthenticator_HappyPath_NotFirstKey(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}
	keyTwo, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
				createCert(keyTwo, 456, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyTwo, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role{
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	res, err := testInstance(context.Background(), jwt)

	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)
}

func TestNewGoogleAuthenticator_HappyPath_Caching(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}
	keyTwo, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	fetchCount := 0
	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		fetchCount++
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
				createCert(keyTwo, 456, start, expires),
			},
			Expiration: time.Now().Add(time.Millisecond * 250),
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())

	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`

	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyOne, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}

	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role{
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)

	res, err := testInstance(context.Background(), jwt)
	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)

	asserter.Equal(1, fetchCount)

	res, err = testInstance(context.Background(), jwt)
	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)
	asserter.Equal(1, fetchCount)

	time.Sleep(time.Millisecond * 500)
	res, err = testInstance(context.Background(), jwt)
	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)
	asserter.Equal(2, fetchCount)
}

func TestNewGoogleAuthenticator_FailureToFetchCertOnFirstTry(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}
	keyTwo, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	shouldBlowUp := true
	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		if shouldBlowUp {
			shouldBlowUp = false
			return GooglePublicCerts{}, errors.New("BWAHAHAAHHA")
		}
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
				createCert(keyTwo, 456, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyTwo, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role{
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), jwt)
	asserter.Error(err)

	res, err := testInstance(context.Background(), jwt)
	asserter.NoError(err)
	asserter.Equal(Principal{
		UserID: subject,
		Email:  email,
		Name:   name,
		Roles:  expectedRoles,
	}, res)
}

func TestNewGoogleAuthenticator_InvalidSigner(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}
	keyTwo, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyTwo, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role {
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), jwt)

	asserter.EqualError(err, fmt.Sprintf("invalid signature on token %s", jwt))
}

func TestNewGoogleAuthenticator_NotJson(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever""", -- to many quotes and stuff
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyOne, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role {
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), jwt)

	asserter.EqualError(err, fmt.Sprintf("garbage token: payload not json (%s)", jwt))
}

func TestNewGoogleAuthenticator_InvalidAudience(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(time.Hour)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s-whoops",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyOne, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role {
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), jwt)

	asserter.EqualError(err, fmt.Sprintf("invalid audience: %s-whoops", clientId))
}

func TestNewGoogleAuthenticator_Expired(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(-time.Second)
	clientId := "testyMcTesterson"
	subject := "12345"
	email := "test@test.com"
	name := "Bob McTester"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	jwtPayload := fmt.Sprintf(`{
  "iss": "foo.bar.com",
  "azp": "whatever",
  "aud": "%s",
  "sub": "%s",
  "email": "%s",
  "email_verified": true,
  "at_hash": "9SrKH5GRtXul1yRpWCyLow",
  "name": "%s",
  "picture": "https://some.google.link/blah",
  "given_name": "Bob",
  "family_name": "McTester",
  "locale": "en",
  "iat": %d,
  "exp": %d,
  "jti": "123abc56"
}`, clientId, subject, email, name, start.Unix(), expires.Unix())
	jwtHeader := `{"alg":"RS256","kid":"a41a3570b8e3ae1b72caabcaa7b8d2db2065d7c1","typ":"JWT"}`
	unsigned := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(jwtHeader)), base64.RawURLEncoding.EncodeToString([]byte(jwtPayload)))
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(unsigned))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, keyOne, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		panic(err)
	}
	jwt := fmt.Sprintf("%s.%s", unsigned, base64.RawURLEncoding.EncodeToString(sigBytes))

	expectedRoles := []Role {
		Role("whatever"),
	}
	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Equal(email, emailAddress)
		return expectedRoles
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), jwt)

	asserter.EqualError(err, fmt.Sprintf("expired token, expiration: %s", expires.Format(time.RFC3339)))
}

func TestNewGoogleAuthenticator_Garbage(t *testing.T) {
	asserter := assert.New(t)

	reader := rand.Reader
	bitSize := 2048

	start := time.Now()
	expires := time.Now().Add(-time.Second)
	clientId := "testyMcTesterson"

	keyOne, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	certFetcher := func(ctx context.Context) (GooglePublicCerts, error) {
		return GooglePublicCerts{
			Certs: []*x509.Certificate{
				createCert(keyOne, 123, start, expires),
			},
			Expiration: expires,
		}, nil
	}

	getRoles := RoleOracle(func(ctx context.Context, emailAddress string) []Role {
		asserter.Fail("should not be here")
		return make([]Role, 0)
	})

	testInstance := NewGoogleAuthenticator(clientId, certFetcher, getRoles)
	_, err = testInstance(context.Background(), "wtfisthis?")
	asserter.EqualError(err, "garbage token: format (wtfisthis?)")

	_, err = testInstance(context.Background(), "YWJj.YWJj.###")
	asserter.EqualError(err, "garbage token: signature malformed (YWJj.YWJj.###)")

	_, err = testInstance(context.Background(), "YWJj.###.YWJj")
	asserter.EqualError(err, "garbage token: payload malformed (YWJj.###.YWJj)")
}

func createCert(key *rsa.PrivateKey, serial int, start time.Time, expires time.Time) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(int64(serial)),
		Subject: pkix.Name{
			Organization: []string{"Testing FTW"},
		},
		NotBefore: start,
		NotAfter:  expires,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	bytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	parsedCert, err := x509.ParseCertificate(bytes)
	if err != nil {
		panic(err)
	}
	return parsedCert
}

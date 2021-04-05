package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/pardot/oidc"
	"github.com/pardot/oidc/signer"
	"gopkg.in/square/go-jose.v2"
)

func TestAuthenticate(t *testing.T) {
	now := time.Now()

	signingKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key: jose.JSONWebKey{
			Key:       testKey,
			KeyID:     "test-key",
			Algorithm: string(jose.RS256),
			Use:       "sig",
		},
	}
	verificationKeys := []jose.JSONWebKey{
		{
			Key:       testKey.Public(),
			KeyID:     "test-key",
			Algorithm: string(jose.RS256),
			Use:       "sig",
		},
	}
	signer := signer.NewStatic(signingKey, verificationKeys)

	cases := []struct {
		name         string
		user         string
		token        string
		userTemplate string
		wantErr      string
	}{
		{
			name: "valid user, valid token",
			user: "jdoe",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"valid-aud"},
				Expiry:    oidc.UnixTime(now.Add(10 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Unix()),
			}),
			wantErr: "",
		},
		{
			name: "valid user, valid token, custom user template",
			user: "jdoe:valid-aud",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"valid-aud"},
				Expiry:    oidc.UnixTime(now.Add(10 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Unix()),
			}),
			userTemplate: "{{.Subject}}:{{index .Audience 0}}",
			wantErr:      "",
		},
		{
			name: "valid user, valid token, invalid custom user template",
			user: "jdoe",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"valid-aud"},
				Expiry:    oidc.UnixTime(now.Add(10 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Unix()),
			}),
			userTemplate: "{{broken}}",
			wantErr:      "parsing user template",
		},
		{
			name: "invalid user, valid token",
			user: "invalid",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"valid-aud"},
				Expiry:    oidc.UnixTime(now.Add(10 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Unix()),
			}),
			wantErr: "expected user \"jdoe\"",
		},
		{
			name: "valid user, expired token",
			user: "jdoe",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"valid-aud"},
				Expiry:    oidc.UnixTime(now.Add(-5 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Add(-5 * time.Minute).Unix()),
			}),
			wantErr: "token is expired",
		},
		{
			name: "valid user, token with incorrect aud",
			user: "jdoe",
			token: mustJWT(t, signer, oidc.Claims{
				Issuer:    "https://example.com",
				Subject:   "jdoe",
				Audience:  []string{"invalid-aud"},
				Expiry:    oidc.UnixTime(now.Add(10 * time.Minute).Unix()),
				NotBefore: oidc.UnixTime(now.Add(-10 * time.Minute).Unix()),
				IssuedAt:  oidc.UnixTime(now.Unix()),
			}),
			wantErr: "invalid audience claim",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			verifier := oidc.NewVerifier("https://example.com", oidc.NewStaticKeysource(jose.JSONWebKeySet{Keys: verificationKeys}))
			auth := &authenticator{
				verifier: verifier,
				aud:      "valid-aud",
			}
			auth.UserTemplate = tc.userTemplate

			err := auth.Authenticate(ctx, tc.user, tc.token)
			if err != nil && tc.wantErr == "" {
				t.Errorf("want no err, got %v", err)
			} else if err != nil && !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want err %v, got %v", tc.wantErr, err)
			} else if err == nil && tc.wantErr != "" {
				t.Errorf("want err %v, got none", tc.wantErr)
			}
		})
	}
}

func mustJWT(t *testing.T, signer *signer.StaticSigner, claims oidc.Claims) string {
	data, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}

	signed, err := signer.Sign(context.TODO(), data)
	if err != nil {
		t.Fatal(err)
	}

	return string(signed)
}

func mustLoadRSAKey(s string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		panic("no pem data found")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

var testKey = mustLoadRSAKey(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEArmoiX5G36MKPiVGS1sicruEaGRrbhPbIKOf97aGGQRjXVngo
Knwd2L4T9CRyABgQm3tLHHcT5crODoy46wX2g9onTZWViWWuhJ5wxXNmUbCAPWHb
j9SunW53WuLYZ/IJLNZt5XYCAFPjAakWp8uMuuDwWo5EyFaw85X3FSMhVmmaYDd0
cn+1H4+NS/52wX7tWmyvGUNJ8lzjFAnnOtBJByvkyIC7HDphkLQV4j//sMNY1mPX
HbsYgFv2J/LIJtkjdYO2UoDhZG3Gvj16fMy2JE2owA8IX4/s+XAmA2PiTfd0J5b4
drAKEcdDl83G6L3depEkTkfvp0ZLsh9xupAvIwIDAQABAoIBABKGgWonPyKA7+AF
AxS/MC0/CZebC6/+ylnV8lm4K1tkuRKdJp8EmeL4pYPsDxPFepYZLWwzlbB1rxdK
iSWld36fwEb0WXLDkxrQ/Wdrj3Wjyqs6ZqjLTVS5dAH6UEQSKDlT+U5DD4lbX6RA
goCGFUeQNtdXfyTMWHU2+4yKM7NKzUpczFky+0d10Mg0ANj3/4IILdr3hqkmMSI9
1TB9ksWBXJxt3nGxAjzSFihQFUlc231cey/HhYbvAX5fN0xhLxOk88adDcdXE7br
3Ser1q6XaaFQSMj4oi1+h3RAT9MUjJ6johEqjw0PbEZtOqXvA1x5vfFdei6SqgKn
Am3BspkCgYEA2lIiKEkT/Je6ZH4Omhv9atbGoBdETAstL3FnNQjkyVau9f6bxQkl
4/sz985JpaiasORQBiTGY8JDT/hXjROkut91agi2Vafhr29L/mto7KZglfDsT4b2
9z/EZH8wHw7eYhvdoBbMbqNDSI8RrGa4mpLpuN+E0wsFTzSZEL+QMQUCgYEAzIQh
xnreQvDAhNradMqLmxRpayn1ORaPReD4/off+mi7hZRLKtP0iNgEVEWHJ6HEqqi1
r38XAc8ap/lfOVMar2MLyCFOhYspdHZ+TGLZfr8gg/Fzeq9IRGKYadmIKVwjMeyH
REPqg1tyrvMOE0HI5oqkko8JTDJ0OyVC0Vc6+AcCgYAqCzkywugLc/jcU35iZVOH
WLdFq1Vmw5w/D7rNdtoAgCYPj6nV5y4Z2o2mgl6ifXbU7BMRK9Hc8lNeOjg6HfdS
WahV9DmRA1SuIWPkKjE5qczd81i+9AHpmakrpWbSBF4FTNKAewOBpwVVGuBPcDTK
59IE3V7J+cxa9YkotYuCNQKBgCwGla7AbHBEm2z+H+DcaUktD7R+B8gOTzFfyLoi
Tdj+CsAquDO0BQQgXG43uWySql+CifoJhc5h4v8d853HggsXa0XdxaWB256yk2Wm
MePTCRDePVm/ufLetqiyp1kf+IOaw1Oyux0j5oA62mDS3Iikd+EE4Z+BjPvefY/L
E2qpAoGAZo5Wwwk7q8b1n9n/ACh4LpE+QgbFdlJxlfFLJCKstl37atzS8UewOSZj
FDWV28nTP9sqbtsmU8Tem2jzMvZ7C/Q0AuDoKELFUpux8shm8wfIhyaPnXUGZoAZ
Np4vUwMSYV5mopESLWOg3loBxKyLGFtgGKVCjGiQvy6zISQ4fQo=
-----END RSA PRIVATE KEY-----`)

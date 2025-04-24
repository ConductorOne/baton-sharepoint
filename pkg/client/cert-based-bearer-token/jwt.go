package certbasedbearertoken

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505 -- SHA-1 required for compatibility reasons with x5t JWT
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

var audienceTemplate = "https://login.microsoftonline.com/%s/v2.0" // tenant ID expected

type JWTOptions struct {
	pfxBase64  string
	password   string
	ClientID   string
	TenantID   string
	TimeUTCNow time.Time
	Duration   time.Duration
	NotBefore  time.Duration
}

func generateSignedJWTFromPFX(opts JWTOptions) (string, error) {
	pfxData, err := base64.StdEncoding.DecodeString(opts.pfxBase64)
	if err != nil {
		return "", fmt.Errorf("cannot decode string with .pfx certificate, error: %w", err)
	}

	privateKey, cert, err := pkcs12.Decode(pfxData, opts.password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt .pfx certificate with password, error: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("key is not RSA")
	}

	nbf := opts.TimeUTCNow.Add(-opts.NotBefore).Unix()
	exp := opts.TimeUTCNow.Add(opts.Duration).Unix()

	// #nosec G401 -- SHA-1 required for thumbprint x5t in JWT
	thumbprint := sha1.Sum(cert.Raw) // x5t is SHA-1 of DER cert
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"x5t": base64UrlEncode(thumbprint[:]),
	}

	payload := map[string]interface{}{
		"aud": fmt.Sprintf(audienceTemplate, opts.TenantID),
		"iss": opts.ClientID,
		"sub": opts.ClientID,
		"jti": fmt.Sprintf("%d-%d", opts.TimeUTCNow.UnixNano(), opts.TimeUTCNow.Unix()),
		"nbf": nbf,
		"exp": exp,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header, error: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload, error: %w", err)
	}

	unsigned := base64UrlEncode(headerJSON) + "." + base64UrlEncode(payloadJSON)

	hashed := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT, error: %w", err)
	}

	signedJWT := unsigned + "." + base64UrlEncode(signature)
	return signedJWT, nil
}

func base64UrlEncode(data []byte) string {
	encoded := strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
	return strings.ReplaceAll(strings.ReplaceAll(encoded, "+", "-"), "/", "_")
}

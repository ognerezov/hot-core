package apple

import (
	"crypto/ecdsa"
	"crypto/x509"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

//go:embed AppleRootCA-G3.cer
var certFS embed.FS

var (
	rootCertPool *x509.CertPool
	initError    error
)

func init() {
	certBytes, err := certFS.ReadFile("AppleRootCA-G3.cer")
	if err != nil {
		initError = fmt.Errorf("failed to read embedded Apple Root CA: %w", err)
		return
	}

	rootCertPool = x509.NewCertPool()
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		initError = fmt.Errorf("failed to parse Apple Root CA: %w", err)
		return
	}
	rootCertPool.AddCert(cert)
}

type TransactionPayload struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	BundleID              string `json:"bundleId"`
	ProductID             string `json:"productId"`
	Environment           string `json:"environment"`
	PurchaseDate          int64  `json:"purchaseDate"`
	OriginalPurchaseDate  int64  `json:"originalPurchaseDate"`
	ExpiresDate           int    `json:"expiresDate"`
	Quantity              int    `json:"quantity"`
	Type                  string `json:"type"`
	InAppOwnershipType    string `json:"inAppOwnershipType"`
	SignedDate            int64  `json:"signedDate"`
	jwt.RegisteredClaims
}

// VerifyTransaction JWS проверяет подпись, валидирует цепочку x5c с Root CA Apple и декодирует payload.
func VerifyTransaction(jwsToken string) (*TransactionPayload, error) {
	if initError != nil {
		return nil, initError
	}

	claims := &TransactionPayload{}

	token, err := jwt.ParseWithClaims(jwsToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Header["alg"])
		}

		x5cRaw, ok := token.Header["x5c"].([]interface{})
		if !ok || len(x5cRaw) == 0 {
			return nil, errors.New("missing or invalid x5c header")
		}

		// 1. Декодируем листовой (основной) сертификат
		certBase64, ok := x5cRaw[0].(string)
		if !ok {
			return nil, errors.New("invalid x5c cert format")
		}

		certBytes, err := base64.StdEncoding.DecodeString(certBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 cert: %w", err)
		}

		leafCert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x509 cert: %w", err)
		}

		// 2. Собираем промежуточные сертификаты из x5c
		intermediates := x509.NewCertPool()
		for i := 1; i < len(x5cRaw); i++ {
			if certStr, ok := x5cRaw[i].(string); ok {
				if b, err := base64.StdEncoding.DecodeString(certStr); err == nil {
					if ic, err := x509.ParseCertificate(b); err == nil {
						intermediates.AddCert(ic)
					}
				}
			}
		}

		// 3. Валидируем цепочку доверия до Apple Root CA
		verifyOpts := x509.VerifyOptions{
			Roots:         rootCertPool,
			Intermediates: intermediates,
		}

		if _, err := leafCert.Verify(verifyOpts); err != nil {
			return nil, fmt.Errorf("apple cert chain verification failed: %w", err)
		}

		// 4. Достаем публичный ключ для проверки подписи JWT
		pubKey, ok := leafCert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not ECDSA")
		}

		return pubKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid transaction token: %w", err)
	}

	return claims, nil
}

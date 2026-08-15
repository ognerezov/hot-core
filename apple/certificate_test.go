package apple

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestChain() (root *x509.Certificate, intermediate *x509.Certificate, leaf *x509.Certificate, leafKey *ecdsa.PrivateKey, err error) {
	// Root CA
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Apple"},
			CommonName:   "Test Apple Root CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	rootDer, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	root, err = x509.ParseCertificate(rootDer)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Intermediate CA
	intKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	intTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Apple"},
			CommonName:   "Test Apple Intermediate CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	intDer, err := x509.CreateCertificate(rand.Reader, intTemplate, root, &intKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	intermediate, err = x509.ParseCertificate(intDer)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Leaf
	leafKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"Test Apple"},
			CommonName:   "Test Apple Leaf",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(1 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}
	leafDer, err := x509.CreateCertificate(rand.Reader, leafTemplate, intermediate, &leafKey.PublicKey, intKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	leaf, err = x509.ParseCertificate(leafDer)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return root, intermediate, leaf, leafKey, nil
}

func TestVerifyTransaction(t *testing.T) {
	rootCert, intermediateCert, leafCert, leafKey, err := generateTestChain()
	if err != nil {
		t.Fatalf("failed to generate test chain: %v", err)
	}

	// Добавляем наш тестовый рут в глобальный пул
	if rootCertPool == nil {
		rootCertPool = x509.NewCertPool()
	}
	rootCertPool.AddCert(rootCert)

	claims := TransactionPayload{
		TransactionID:         "1000000000000001",
		OriginalTransactionID: "1000000000000001",
		BundleID:              "com.example.app",
		ProductID:             "com.example.product1",
		Environment:           "Sandbox",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	t.Run("Valid Token with Chain", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["x5c"] = []interface{}{
			base64.StdEncoding.EncodeToString(leafCert.Raw),
			base64.StdEncoding.EncodeToString(intermediateCert.Raw),
		}
		tokenString, err := token.SignedString(leafKey)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		payload, err := VerifyTransaction(tokenString)
		if err != nil {
			t.Fatalf("VerifyTransaction failed: %v", err)
		}

		if payload.TransactionID != claims.TransactionID {
			t.Errorf("expected transaction ID %s, got %s", claims.TransactionID, payload.TransactionID)
		}
		if payload.BundleID != claims.BundleID {
			t.Errorf("expected bundle ID %s, got %s", claims.BundleID, payload.BundleID)
		}
		if payload.ProductID != claims.ProductID {
			t.Errorf("expected product ID %s, got %s", claims.ProductID, payload.ProductID)
		}
		if payload.Environment != claims.Environment {
			t.Errorf("expected environment %s, got %s", claims.Environment, payload.Environment)
		}
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["x5c"] = []interface{}{
			base64.StdEncoding.EncodeToString(leafCert.Raw),
		}

		// Генерируем другой ключ для невалидной подписи
		otherKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		tokenString, _ := token.SignedString(otherKey)

		_, err := VerifyTransaction(tokenString)
		if err == nil {
			t.Error("expected error for invalid signature, got nil")
		}
	})

	t.Run("Missing x5c", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		tokenString, _ := token.SignedString(leafKey)

		_, err := VerifyTransaction(tokenString)
		if err == nil {
			t.Error("expected error for missing x5c, got nil")
		}
	})

	t.Run("Untrusted Root", func(t *testing.T) {
		// Генерируем другую цепочку, рут которой не добавлен в rootCertPool
		_, _, otherLeafCert, otherLeafKey, _ := generateTestChain()

		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["x5c"] = []interface{}{
			base64.StdEncoding.EncodeToString(otherLeafCert.Raw),
		}
		tokenString, _ := token.SignedString(otherLeafKey)

		_, err := VerifyTransaction(tokenString)
		if err == nil {
			t.Error("expected error for untrusted root, got nil")
		}
	})
}

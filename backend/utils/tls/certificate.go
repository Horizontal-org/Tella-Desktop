package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	certFileName = "cert.pem"
	keyFileName  = "key.pem"
)

func LoadCertificate(certPath string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

func GetCertificatePaths() (certPath, keyPath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	tellaDir := filepath.Join(homeDir, ".tella")
	err = os.MkdirAll(tellaDir, 0700)
	if err != nil {
		return "", "", err
	}

	certPath = filepath.Join(tellaDir, certFileName)
	keyPath = filepath.Join(tellaDir, keyFileName)
	return certPath, keyPath, nil
}

// GenerateCertificate creates a new self-signed certificate and private key
func GenerateCertificate() ([]byte, []byte, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Prepare certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Tella Desktop"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate
	certOut := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	// Encode private key
	keyOut := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certOut, keyOut, nil
}

// SaveCertificateAndKey saves the certificate and private key to disk
func SaveCertificateAndKey(cert, key []byte) error {
	certPath, keyPath, err := GetCertificatePaths()
	if err != nil {
		return err
	}

	if err := os.WriteFile(certPath, cert, 0600); err != nil {
		return err
	}

	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return err
	}

	return nil
}

// LoadCertificateAndKey loads the certificate and private key from disk
// If they don't exist, generates new ones and saves them
func LoadOrGenerateCertificateAndKey() (certPath string, keyPath string, err error) {
	certPath, keyPath, err = GetCertificatePaths()
	if err != nil {
		return "", "", err
	}

	// Check if files exist
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)

	// If either file doesn't exist, generate new ones
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		cert, key, err := GenerateCertificate()
		if err != nil {
			return "", "", err
		}

		if err := SaveCertificateAndKey(cert, key); err != nil {
			return "", "", err
		}
	}

	return certPath, keyPath, nil
}

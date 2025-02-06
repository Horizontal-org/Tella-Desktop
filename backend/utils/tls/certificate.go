package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
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

type Config struct {
	CommonName   string
	Organization []string
	IPAddresses  []net.IP
}

type certificateFiles struct {
	certPath string
	keyPath  string
}

// generates a TLS configuration with a self signed certificate
func GenerateTLSConfig(config Config) (*tls.Config, error) {
	cert, err := generateCertificate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

// creates self signed certificate and returns it as a tls certificate
func generateCertificate(config Config) (tls.Certificate, error) {
	// Generate private key
	privateKey, err := generatePrivateKey()
	if err != nil {
		return tls.Certificate{}, err
	}

	template, err := createCertificateTemplate(config)
	if err != nil {
		return tls.Certificate{}, err
	}

	derBytes, err := createSignedCertificate(template, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	files, err := setupCertificateFiles(derBytes, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(files.certPath, files.keyPath)

}

func generatePrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return privateKey, nil
}

func createCertificateTemplate(config Config) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: config.Organization,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           config.IPAddresses,
	}, nil
}

func createSignedCertificate(template *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	return derBytes, nil
}

func setupCertificateFiles(derBytes []byte, privateKey *rsa.PrivateKey) (*certificateFiles, error) {
	certDir := getCertificateDirectory()
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %w", err)
	}

	files := &certificateFiles{
		certPath: filepath.Join(certDir, "tella.crt"),
		keyPath:  filepath.Join(certDir, "tella.key"),
	}

	if err := writeCertificateFile(files.certPath, derBytes); err != nil {
		return nil, err
	}

	if err := writePrivateKeyFile(files.keyPath, privateKey); err != nil {
		return nil, err
	}

	return files, nil
}

func writeCertificateFile(path string, derBytes []byte) error {
	certOut, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}
	return nil
}

func writePrivateKeyFile(path string, privateKey *rsa.PrivateKey) error {
	keyOut, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}
	return nil
}

// returns the directory where certificates should be stored
func getCertificateDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "certs")
	}
	return filepath.Join(homeDir, "Documents", "TellaCerts")
}

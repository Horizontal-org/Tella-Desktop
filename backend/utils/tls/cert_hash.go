package tls

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// GetCertificateHash generates a human-readable hash from a certificate
func GetCertificateHash(cert *x509.Certificate) string {
	// Get the raw certificate data and hash it
	hash := sha256.Sum256(cert.Raw)
	// Take first 8 bytes and encode them to base64
	return base64.StdEncoding.EncodeToString(hash[:8])
}

// GetCertificateDisplayString formats certificate information for display
func GetCertificateDisplayString(cert *x509.Certificate) string {
	hash := GetCertificateHash(cert)
	return fmt.Sprintf("Certificate Fingerprint: %s\nPlease verify this matches on both devices", hash)
}

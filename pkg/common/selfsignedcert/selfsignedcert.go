// SPDX-License-Identifier: GPL-3.0-or-later

// Package selfsignedcert helps to create self-signed certificates.
package selfsignedcert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/bassosimone/runtimex"
)

// Config contains configuration for [New].
type Config struct {
	// CommonName is the certificate common name.
	CommonName string

	// DNSNames contains the alternative DNS names to include in the certificate.
	DNSNames []string

	// IPAddrs contains the IP addrs for which the certificate is valid.
	IPAddrs []net.IP
}

// NewConfigExampleCom creates a [*Config] for example.com
// using the www.example.com, 127.0.0.1, and ::1 sans.
func NewConfigExampleCom() *Config {
	config := &Config{
		CommonName: "example.com",
		DNSNames:   []string{"www.example.com"},
		IPAddrs: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}
	return config
}

// Cert is the self-signed certificate.
type Cert struct {
	// CertPEM is the certificate encoded using PEM.
	CertPEM []byte

	// KeyPEM is the secret key encoded using PEM.
	KeyPEM []byte
}

// WriteFiles writes CertPEM to `cert.pem` and KeyPEM to `key.pem`.
//
// This method panics on failure.
func (c *Cert) WriteFiles(baseDir string) {
	runtimex.PanicOnError0(os.WriteFile(filepath.Join(baseDir, "cert.pem"), c.CertPEM, 0600))
	runtimex.PanicOnError0(os.WriteFile(filepath.Join(baseDir, "key.pem"), c.KeyPEM, 0600))
}

// New generates a self-signed certificate and key with SANs.
//
// This function panics on failure.
func New(config *Config) *Cert {
	// Generate the private key
	priv := runtimex.PanicOnError1(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))

	// Build the certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber := runtimex.PanicOnError1(rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"RBMK Project"},
			CommonName:   config.CommonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs to the certificate
	template.DNSNames = config.DNSNames
	template.IPAddresses = config.IPAddrs

	// Generate the certificate proper and encoded to PEM
	certDER := runtimex.PanicOnError1(x509.CreateCertificate(
		rand.Reader, &template, &template, &priv.PublicKey, priv))
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Generate the private key in PEM format
	keyPEM := runtimex.PanicOnError1(x509.MarshalECPrivateKey(priv))
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyPEM})

	// Return the results
	return &Cert{CertPEM: certPEM, KeyPEM: keyPEMBytes}
}

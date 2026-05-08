package shttp

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"
	"time"
)

// generateCertificate contains all inputs for runtime self-signed certificate generation.
type generateCertificate struct {
	ServiceName  string
	Organization string
	Host         string
	DNSnames     []string
	IPs          []string
	ValidFrom    string
	ValidFor     time.Duration
	IsCA         bool
	RSABits      int
	EcdsaCurve   string
	Ed25519Key   bool
}

func (gc *generateCertificate) publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

// GeneratePEM generates a self-signed certificate and key in PEM format.
func (gc *generateCertificate) GeneratePEM() (certPEM []byte, keyPEM []byte, err error) {
	var priv any

	switch gc.EcdsaCurve {
	case "":
		if gc.Ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			bits := gc.RSABits
			if bits == 0 {
				bits = 2048
			}
			priv, err = rsa.GenerateKey(rand.Reader, bits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, nil, fmt.Errorf("unrecognized elliptic curve: %q", gc.EcdsaCurve)
	}
	if err != nil {
		return nil, nil, err
	}

	var notBefore time.Time
	if len(gc.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", gc.ValidFrom)
		if err != nil {
			return nil, nil, err
		}
	}

	template, err := gc.createTemplate(notBefore)
	if err != nil {
		return nil, nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, gc.publicKey(priv), priv)
	if err != nil {
		return nil, nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	return certPEM, keyPEM, nil
}

// GenerateTLSConfig generates a TLS config with a self-signed certificate.
func (gc *generateCertificate) GenerateTLSConfig() (*tls.Config, error) {
	certPEM, keyPEM, err := gc.GeneratePEM()
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}, nil
}

// GenerateCertificateFromConfig generates a self-signed certificate and key
// based on the HTTP service config and returns them as PEM bytes.
func GenerateCertificateFromConfig(cfg Config) (certPEM []byte, keyPEM []byte, err error) {
	host := cfg.ServiceURL
	if host == "" {
		host = "localhost"
	} else if u, parseErr := url.Parse(host); parseErr == nil {
		if parsedHost := u.Hostname(); parsedHost != "" {
			host = parsedHost
		}
	}

	gc := generateCertificate{
		ServiceName:  cfg.Servicename,
		Host:         host,
		ValidFor:     10 * 365 * 24 * time.Hour,
		IsCA:         false,
		EcdsaCurve:   "P384",
		Ed25519Key:   false,
		DNSnames:     cfg.DNSNames,
		IPs:          cfg.IPAddresses,
		Organization: cfg.Servicename,
	}

	return gc.GeneratePEM()
}

func (gc *generateCertificate) createTemplate(notBefore time.Time) (*x509.Certificate, error) {
	notAfter := notBefore.Add(gc.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{gc.Organization},
			CommonName:   gc.ServiceName,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, sip := range gc.IPs {
		if ip := net.ParseIP(sip); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	template.DNSNames = append(template.DNSNames, gc.DNSnames...)

	hosts := strings.Split(gc.Host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if gc.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	return &template, nil
}

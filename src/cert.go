package src

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func PrepareCertifications() {
	if _, err := os.Stat(SSLDir); os.IsNotExist(err) {
		log.Printf("[*] creating SSL directory at %s", SSLDir)
		if err := os.MkdirAll(SSLDir, 0755); err != nil {
			log.Fatalf("[-] failed to create SSL directory: %v", err)
		}
	}

	certExists := true
	privateKeyExists := true

	if _, err := os.Stat(CertPath); os.IsNotExist(err) {
		certExists = false
	}

	if _, err := os.Stat(PrivateKeyPath); os.IsNotExist(err) {
		privateKeyExists = false
	}

	if !certExists && !privateKeyExists {
		log.Printf("[*] generating new self-signed certificate")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("[-] failed to generate private key: %v", err)
		}

		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				Organization: []string{"CreemProxy Self Signed Cert"},
				CommonName:   "CreemProxy",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(1000, 0, 0),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
			IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			log.Fatalf("[-] failed to create certificate: %v", err)
		}

		privateKeyFile, err := os.Create(PrivateKeyPath)
		if err != nil {
			log.Fatalf("[-] failed to create private key file: %v", err)
		}
		defer privateKeyFile.Close()

		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}
		if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
			log.Fatalf("[-] failed to write private key: %v", err)
		}

		certFile, err := os.Create(CertPath)
		if err != nil {
			log.Fatalf("[-] failed to create certificate file: %v", err)
		}
		defer certFile.Close()

		certPEM := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		}
		if err := pem.Encode(certFile, certPEM); err != nil {
			log.Fatalf("[-] failed to write certificate: %v", err)
		}

		log.Printf("[+] new self-signed certificate generated")
	} else {
		log.Printf("[*] using existing SSL certificate")
	}

	certData, err := os.ReadFile(CertPath)
	if err != nil {
		log.Fatalf("[-] failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		log.Fatalf("[-] failed to parse PEM block containing the certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("[-] failed to parse certificate: %v", err)
	}

	sha1sum := fmt.Sprintf("%X", sha1.Sum(cert.Raw))
	log.Printf("[+] certificate fingerprint (sha1): %s", sha1sum)
}

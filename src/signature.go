package src

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

func PrepareSigningAsset() {
	if _, err := os.Stat(SigningDir); os.IsNotExist(err) {
		log.Printf("[*] creating signing key directory at %s", SigningDir)
		if err := os.MkdirAll(SigningDir, 0755); err != nil {
			log.Fatalf("[-] failed to create signing key directory: %v", err)
		}
	}

	publicKeyExists := true
	privateKeyExists := true

	if _, err := os.Stat(SigningPublicKeyPath); os.IsNotExist(err) {
		publicKeyExists = false
	}

	if _, err := os.Stat(SigningPrivateKeyPath); os.IsNotExist(err) {
		privateKeyExists = false
	}

	if !publicKeyExists && !privateKeyExists {
		log.Printf("[*] generating new Ed25519 signing key pair")

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatalf("[-] failed to generate Ed25519 key pair: %v", err)
		}

		// Save public key
		if err := os.WriteFile(SigningPublicKeyPath, publicKey, 0644); err != nil {
			log.Fatalf("[-] failed to write public key: %v", err)
		}
		log.Printf("[+] public key saved to %s", SigningPublicKeyPath)

		// Save private key
		if err := os.WriteFile(SigningPrivateKeyPath, privateKey, 0600); err != nil {
			log.Fatalf("[-] failed to write private key: %v", err)
		}
		log.Printf("[+] private key saved to %s", SigningPrivateKeyPath)

		log.Printf("[+] new Ed25519 signing key pair generated and saved")
	} else if !publicKeyExists {
		log.Fatalf("[-] public signing key missing, but private key exists. Please resolve this inconsistency.")
	} else if !privateKeyExists {
		log.Fatalf("[-] private signing key missing, but public key exists. Please resolve this inconsistency.")
	} else {
		log.Printf("[*] using existing Ed25519 signing key pair")
	}

	publicKeyBytes, err := os.ReadFile(SigningPublicKeyPath)
	if err != nil {
		log.Fatalf("[-] failed to read public signing key: %v", err)
	}
	log.Printf("[+] signing public key (base64): %s", base64.StdEncoding.EncodeToString(publicKeyBytes))
}

func Sign(data []byte) (string, error) {
	privateKeyBytes, err := os.ReadFile(SigningPrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private signing key: %w", err)
	}
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("private key has incorrect size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)
	signature := ed25519.Sign(privateKey, data)
	return base64.StdEncoding.EncodeToString(signature), nil
}

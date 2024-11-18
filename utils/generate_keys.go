package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

const PEM_FOLDER = "pem"

// Source: https://systemweakness.com/generating-rsa-pem-key-pair-using-go-7fd9f1471b58
func GenerateKeyPair(prefix string) ([]byte, []byte) {
	os.MkdirAll(PEM_FOLDER, os.ModePerm)

	prefix = fmt.Sprintf(`%s/%s`, PEM_FOLDER, prefix)

	// Generate a new RSA private key with 2048 bits
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating RSA private key:", err)
		os.Exit(1)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyPath := fmt.Sprintf(`%s_private_key.pem`, prefix)
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		fmt.Println("Error creating private key file:", err)
		os.Exit(1)
	}
	pem.Encode(privateKeyFile, privateKeyPEM)
	privateKeyFile.Seek(0, 0) //seek to the start of file or else ReadAll will result with empty byte slice
	privateKeyBytes, err := io.ReadAll(privateKeyFile)
	defer privateKeyFile.Close()

	// Extract the public key from the private key
	publicKey := &privateKey.PublicKey

	// Encode the public key to the PEM format
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	}

	publicKeyPath := fmt.Sprintf(`%s_public_key.pem`, prefix)
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		fmt.Println("Error creating public key file:", err)
		os.Exit(1)
	}
	pem.Encode(publicKeyFile, publicKeyPEM)
	publicKeyFile.Seek(0, 0) //seek to the start of file or else ReadAll will result with empty byte slice
	publicKeyBytes, err := io.ReadAll(publicKeyFile)

	defer publicKeyFile.Close()

	fmt.Println("RSA key pair generated successfully!")

	return publicKeyBytes, privateKeyBytes // (public, private)
}

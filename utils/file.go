package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const MAX_READ_BUFFER = 2 * 1024 * 1024 // 2 MB

func signSignature(filePath string, privateKeyBytes []byte) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return nil, errors.New("invalid public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing public key")
	}

	buf := make([]byte, MAX_READ_BUFFER)

	checksum := sha256.New()

	for {
		n, err := file.Read(buf)
		if err != nil || n == 0 {
			break
		}
		
		checksum.Write(buf)
	}

	hash := checksum.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func verifySignature(filePath string, signature []byte, publicKeyBytes []byte) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		return nil, errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKCS8PrivateKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing public key")
	}

	buf := make([]byte, MAX_READ_BUFFER)

	checksum := sha256.New()

	for {
		n, err := file.Read(buf)
		if err != nil || n == 0 {
			break
		}
		
		checksum.Write(buf)
	}

	hash := checksum.Sum(nil)

	err = rsa.VerifyPSS(publicKey.(*rsa.PublicKey), crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return nil, err
	}

	return signature, nil
}



func encryptFile(filePath string, publicKeyBytes []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		return errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return errors.New("error parsing public key")
	}

	encryptedFile, err := os.Create(fmt.Sprintf(`%s-encrypted.txt`, filepath.Base(filePath)))
	if err != nil {
		return err
	}
	defer encryptedFile.Close()

	buf := make([]byte, MAX_READ_BUFFER)

	for {
		n, err := file.Read(buf)
		if err != nil || n == 0 {
			break
		}
		encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), buf)
		_, err = encryptedFile.Write(encryptedData)
		if err != nil {
			break
		}
	}

	return nil
}

func decryptFile(filePath string, privateKeyBytes []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return errors.New("invalid public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return errors.New("error parsing public key")
	}

	encryptedFile, err := os.Create(fmt.Sprintf(`%s-decrypted.txt`, filepath.Base(filePath)))
	if err != nil {
		return err
	}
	defer encryptedFile.Close()

	buf := make([]byte, MAX_READ_BUFFER)

	for {
		n, err := file.Read(buf)
		if err != nil || n == 0 {
			break
		}
		decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), buf)
		_, err = encryptedFile.Write(decryptedData)
		if err != nil {
			break
		}
	}

	return nil
}

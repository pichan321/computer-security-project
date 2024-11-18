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

	"github.com/google/uuid"
)

const (
	RSA_KEY_SIZE            = 256
	RSA_PADDING_OVERHEAD    = 42
	RSA_MAX_ENCRYPTION_SIZE = RSA_KEY_SIZE - RSA_PADDING_OVERHEAD
)
const MAX_READ_BUFFER = 32

func SignSignature(filePath string, privateKeyBytes []byte) ([]byte, error) {
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

func VerifySignature(filePath string, signature []byte, publicKeyBytes []byte) ([]byte, error) {
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

func EncryptFile(filePath string, publicKeyBytes []byte) (string, error) {
	uuid := uuid.New().String()

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		return "", errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return "", errors.New("error parsing public key")
	}

	encryptedFileName := fmt.Sprintf(`%s%s`, uuid, filepath.Ext(filePath))
	encryptedFile, err := os.Create(encryptedFileName)
	if err != nil {
		return "", err
	}
	defer encryptedFile.Close()

	buf := make([]byte, RSA_MAX_ENCRYPTION_SIZE)

	for {
		n, err := file.Read(buf)
		if n == 0 || err != nil {
			break
		}

		encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, buf[:n])
		if err != nil {
			return "", err
		}

		_, err = encryptedFile.Write(encryptedData)
		if err != nil {
			return "", err
		}
	}

	return encryptedFileName, nil
}

func DecryptFile(filePath string, privateKeyBytes []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return errors.New("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return errors.New("error parsing private key")
	}

	decryptedFile, err := os.Create(fmt.Sprintf(`%s-decrypted`, filepath.Base(filePath)))
	if err != nil {
		return err
	}
	defer decryptedFile.Close()

	buf := make([]byte, RSA_KEY_SIZE)

	for {
		n, err := file.Read(buf)
		if n == 0 || err != nil {
			break
		}

		decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, buf[:n])
		if err != nil {
			return err
		}

		_, err = decryptedFile.Write(decryptedData)
		if err != nil {
			return err
		}
	}

	return nil
}

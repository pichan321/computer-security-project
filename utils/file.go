package utils

import (
	"crypto"
	"crypto/md5"
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
		return nil, errors.New("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing private key")
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

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
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

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
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

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func EncryptKey(keyToBeEncryptedBytes []byte, publicKeyBytes []byte) ([]byte, error) {
	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		return nil, errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing public key")
	}

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, keyToBeEncryptedBytes)
	if err != nil {
		return nil, errors.New("unexpected error encrypting key")
	}

	return encryptedKey, nil
}

func DecryptKey(encryptedKeyToBeDecryptedBytes []byte, privateKeyBytes []byte) ([]byte, error) {
	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return nil, errors.New("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing private key")
	}

	decryptedKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedKeyToBeDecryptedBytes)
	if err != nil {
		return nil, err
	}

	return decryptedKey, nil
}

func EncryptFile(filePath string, publicKeyBytes []byte) (string, string, error) {
	uuid := uuid.New().String()

	checksum := md5.New()
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		return "", "", errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return "", "", errors.New("error parsing public key")
	}

	encryptedFileName := fmt.Sprintf(`%s%s`, uuid, filepath.Ext(filePath))
	encryptedFile, err := os.Create(encryptedFileName)
	if err != nil {
		return "", "", err
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
			return "", "", err
		}

		_, err = encryptedFile.Write(encryptedData)
		checksum.Write(buf[:n])

		if err != nil {
			return "", "", err
		}
	}

	checksumHash := string(checksum.Sum(nil))

	return encryptedFileName, checksumHash, nil
}

func DecryptFile(filePath string, privateKeyBytes []byte) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return "", "", errors.New("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return "", "", errors.New("error parsing private key")
	}
 
	decryptedFilePath := fmt.Sprintf(`%s-decrypted`, filepath.Base(filePath))
	decryptedFile, err := os.Create(decryptedFilePath)
	if err != nil {
		return "", "", err
	}
	defer decryptedFile.Close()

	buf := make([]byte, RSA_KEY_SIZE)

	checksum := md5.New()
	for {
		n, err := file.Read(buf)
		if n == 0 || err != nil {
			break
		}

		decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, buf[:n])
		checksum.Write(buf[:n])
		if err != nil {
			return "", "", err
		}

		_, err = decryptedFile.Write(decryptedData)
		if err != nil {
			return "", "", err
		}
	}

	checksumHash := string(checksum.Sum(nil))
	return decryptedFilePath, checksumHash, nil
}

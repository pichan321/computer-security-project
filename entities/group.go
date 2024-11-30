package entities

import (
	"blockchain-fileshare/ipfs"
	keys "blockchain-fileshare/keys"
	"blockchain-fileshare/utils"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

// this struct is to make life easy to deal with Member interface change
type Operators struct {
	proxy      *IPFSProxy
	sh         *shell.Shell
	blockchain *Blockchain
}

type Member interface {
	IsMember() bool
	IsOwner() bool
	GetUuid() string
	GetPublicKey() []byte
}

type File struct {
	fileExtension string
	fileOwner     Member
	FileName      string //this is probably what the users will ever see on the interface
	handle        string //this is not actually, necessary for the system to work, but it is required for testing the security later
	transactionID string
}

type Group struct {
	groupID      string
	groupMembers []Member
	files        []File
}

type UserMetadata struct { //this is like GroupMember/GroupOwner but since we don't want private key to be stored in the proxy, I chose to go with this struct
	uuid      string
	publicKey []byte
}

// for the distributed access control policies and group key management (Huang et al.)
/**
 I just realized that the paper uses one hash table to store the group keys (public + private) and
all members' public keys (each user should never have to give out their private keys)

 My approach stores the group's public and private as fields, and all users/members separately in a slice just so that
I could easy print just the users and debug stuff. It comes with a worse runtime lookup but better and clearer birdeye view
of what's going on for anyone looking through my code
**/

type GroupMetadata struct {
	ownerUuid  string
	groupUuid  string //this might be redundant but we might need it later
	publicKey  []byte
	privateKey []byte
	users      []UserMetadata
}

type IPFSProxy struct {
	groups map[string]GroupMetadata
}

type UploadRequest struct {
	filePath          string
	groupID           string
	requestedUserUuid string
	signature         []byte
}

type DownloadRequest struct {
	requestedUserId string
	groupId         string
	IPFSHandle      string
	fileExtension   string
}

func encodeDownloadRequest(downloadRequest DownloadRequest) ([]byte, error) {
	var downloadReqStructBytesBuffer bytes.Buffer
	encoder := gob.NewEncoder(&downloadReqStructBytesBuffer)
	err := encoder.Encode(downloadRequest)
	if err != nil {
		return nil, err
	}

	return downloadReqStructBytesBuffer.Bytes(), nil
}

func SignDownloadRequest(downloadRequest DownloadRequest, privateKeyBytes []byte) ([]byte, error) {
	downloadRequestBytes, err := encodeDownloadRequest(downloadRequest)
	if err != nil {
		return nil, err
	}

	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		return nil, errors.New("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing private key")
	}

	checksum := sha256.New()
	checksum.Write(downloadRequestBytes)
	hash := checksum.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

//return slice of old files for testing our threat model
func (proxy *IPFSProxy) ChangeKeyAndSecureFiles(operator *Operators, groupOwner *GroupOwner, groupIdx int, groupID string) ([]File, error) {
	group, exists := proxy.groups[groupID]
	if !exists {
		return nil, errors.New("group does not exist")
	}

	public, private := keys.GenerateKeyPair(group.groupUuid)

	groupMetadata := groupOwner.groupsOwned[groupIdx]
	oldFiles := groupMetadata.files
	newFilePaths := []string{}
	for _, file := range groupMetadata.files {
		decryptedFilePath, _, err := groupOwner.DownloadFile(operator, groupID, file.transactionID) //verifying checksum is optional, we already know that all the files (are not tampered)/(have not been tampered)
		if err != nil {
			continue
		}
		newFilePaths = append(newFilePaths, decryptedFilePath)
		err = ipfs.DeleteFileFromIPFS(operator.sh, file.handle)
		if err != nil {
			continue
		}
	}

	group.publicKey = public
	group.privateKey = private
	proxy.groups[groupID] = group

	groupOwner.groupsOwned[groupIdx].files = []File{}

	for _, newFilePath := range newFilePaths {
		newFilePath, _ = strings.CutSuffix(newFilePath, "-decrypted")
		transactionID, err := groupOwner.UploadFile(operator, groupID, newFilePath)
		if err != nil {
			continue
		}
		fmt.Println("NEW TRANS", transactionID)

	}

	fmt.Println(groupOwner.groupsOwned[groupIdx].files)

	return oldFiles, nil
}

func (proxy IPFSProxy) getGroupPrivateKey(groupID string) ([]byte, error) {
	group, ok := proxy.groups[groupID]
	if !ok {
		return nil, errors.New("group does not exist")
	}

	return group.privateKey, nil
}

func (proxy IPFSProxy) VerifyDownloadReqSignature(downloadRequest DownloadRequest, signature []byte) ([]byte, error) {
	requestedUserPublicKey, err := proxy.getUserPublicKey(downloadRequest.groupId, downloadRequest.requestedUserId)
	if err != nil {
		return nil, err
	}

	publicKeyBlock, _ := pem.Decode(requestedUserPublicKey)
	if publicKeyBlock == nil {
		return nil, errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing public key")
	}

	downloadRequestBytes, err := encodeDownloadRequest(downloadRequest)
	if err != nil {
		return nil, err
	}

	checksum := sha256.New()
	checksum.Write(downloadRequestBytes)
	hash := checksum.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (proxy IPFSProxy) DownloadFileFromIPFS(sh *shell.Shell, downloadRequest DownloadRequest) (string, []byte, error) {
	err := ipfs.DownloadFileFromIPFS(sh, downloadRequest.IPFSHandle, downloadRequest.fileExtension)
	if err != nil {
		return "", nil, err
	}

	encryptedFileName := downloadRequest.IPFSHandle + downloadRequest.fileExtension
	groupPrivateKey, err := proxy.getGroupPrivateKey(downloadRequest.groupId)
	if err != nil {
		return "", nil, err
	}
	return encryptedFileName, groupPrivateKey, nil

}

func (proxy IPFSProxy) getUserPublicKey(groupID string, uuid string) ([]byte, error) {
	group, ok := proxy.groups[groupID]
	if !ok {
		return nil, errors.New("group does not exist")
	}
	for _, m := range group.users {
		if m.uuid == uuid {
			return m.publicKey, nil
		}
	}
	return nil, errors.New("user is not a member of the group")
}

func (proxy IPFSProxy) getGroupPublicKey(groupID string) ([]byte, error) {
	group, ok := proxy.groups[groupID]
	if !ok {
		return nil, errors.New("group does not exist")
	}

	return group.publicKey, nil
}

// in real world, filePath would be the actual file instead
func (proxy IPFSProxy) VerifySignature(signature []byte, uploadReq UploadRequest) error {
	requestedUserPublicKey, err := proxy.getUserPublicKey(uploadReq.groupID, uploadReq.requestedUserUuid)
	if err != nil {
		return err
	}

	_, err = utils.VerifySignature(uploadReq.filePath, signature, requestedUserPublicKey)
	if err != nil {
		return err
	}
	return nil
}

func (proxy IPFSProxy) UploadFileToIPFS(sh *shell.Shell, uploadReq UploadRequest) (string, string, error) {
	groupPublicKey, err := proxy.getGroupPublicKey(uploadReq.groupID)
	if err != nil {
		return "", "", err
	}

	handle, checksum, err := ipfs.UploadFileToIPFS(sh, uploadReq.filePath, groupPublicKey)

	return handle, checksum, nil
}

func (proxy IPFSProxy) PrintUsers(groupID string) {
	for _, m := range proxy.groups[groupID].users {
		fmt.Println(m.uuid)
	}
}

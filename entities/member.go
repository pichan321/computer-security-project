package entities

import (
	"blockchain-fileshare/utils"
	"errors"
	"fmt"
	"path/filepath"
)

type GroupMember struct {
	uuid       string
	publicKey  []byte
	privateKey []byte
}

func (g GroupMember) IsMember() bool {
	return true
}

func (g GroupMember) IsOwner() bool {
	return false
}

func (g GroupMember) GetPublicKey() []byte {
	return g.publicKey
}

func (g GroupMember) GetUuid() string {
	return g.uuid
}

func (g GroupMember) IsMemberOf(proxy *IPFSProxy, groupID string) (bool, error) {
	for _, m := range proxy.groups[groupID].users {
		if m.uuid == g.GetUuid() {
			return true, nil
		}
	}

	return false, errors.New("is not a member")
}

// this function is necessary because private key of each GroupMember should not be exposed by any means
func (g GroupMember) SignSignature(filePath string) ([]byte, error) {
	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (g *GroupMember) UploadFile(operator *Operators, groupOwner *GroupOwner, groupID string, filePath string) (string, error) {
	if isMember, err := g.IsMemberOf(operator.proxy, groupID); !isMember {
		return "", err
	}

	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
		fmt.Println("here")
		return "", err
	}

	uploadReq := UploadRequest{
		filePath:          filePath,
		groupID:           groupID,
		requestedUserUuid: g.GetUuid(),
		signature:         signature,
	}

	err = operator.proxy.VerifySignature(signature, uploadReq)
	if err != nil {
		return "", err
	}

	//a good implementation should have VerifySignature wrapped inside UploadFileToIPFS
	handle, checksum, err := operator.proxy.UploadFileToIPFS(operator.sh, uploadReq)
	if err != nil {
		return "", err
	}

	transactionData := Data{
		userId:        g.GetUuid(),
		groupId:       groupID,
		fileHash:      checksum,
		IPFSHash:      handle,
		fileExtension: filepath.Ext(filePath),
	}
	transactionHash := operator.blockchain.CreateTransaction(transactionData)

	file := File{
		fileExtension: filepath.Ext(filePath),
		fileOwner:     g,
		FileName:      filepath.Base(filePath),
		handle:        handle,
		transactionID: transactionHash,
	}

	groupIdx := -1
	for idx, group := range groupOwner.groupsOwned {
		if group.groupID == groupID {
			groupIdx = idx
			break
		}
	}
	if groupIdx == -1 {
		return "", errors.New("unexpected error while finding group to insert the uploaded file metadata into")
	}

	groupOwner.groupsOwned[groupIdx].files = append(groupOwner.groupsOwned[groupIdx].files, file)
	return transactionHash, nil
}

func (g GroupMember) ReadFile(operator *Operators, groupID string, handle string) error {

	return nil
}

func (g GroupMember) DownloadFile(operator *Operators, groupID string, transactionHash string) (string, string, error) {
	data, err := operator.blockchain.GetTransactionByHash(transactionHash)
	if err != nil {
		return "", "", nil
	}

	downloadRequest := DownloadRequest{
		requestedUserId: g.GetUuid(),
		groupId:         groupID,
		IPFSHandle:      data.IPFSHash,
	}

	signature, err := SignDownloadRequest(downloadRequest, g.privateKey)
	if err != nil {
		return "", "", err
	}

	_, err = operator.proxy.VerifyDownloadReqSignature(downloadRequest, signature)
	if err != nil {
		return "", "", err
	}

	file, groupPrivateKey, err := operator.proxy.DownloadFileFromIPFS(operator.sh, downloadRequest)
	if err != nil {
		return "", "", err
	}

	decryptedFilePath, checksumHash, err := utils.DecryptFile(file, groupPrivateKey)
	if err != nil {
		return "", "", err
	}

	return decryptedFilePath, checksumHash, nil
}

func (g GroupMember) DeleteFile(operator *Operators, groupID string, handle string) error {

	return nil
}

func CreateIPFSProxy() *IPFSProxy {
	return &IPFSProxy{
		groups: map[string]GroupMetadata{},
	}
}

func isValidMember(memberUuid string, allUsers []Member) (Member, bool) {
	for _, member := range allUsers {
		if member.GetUuid() == memberUuid {
			return member, true
		}
	}

	return nil, false
}

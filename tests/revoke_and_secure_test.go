package tests

import (
	"blockchain-fileshare/entities"
	"blockchain-fileshare/ipfs"
	"blockchain-fileshare/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TEST_FILEPATH = "sample.txt"
const GROUP_ONE_MEMBER_COUNT = 1

func TestRevokeAndSecureFiles(t *testing.T) {
	groupOwner := entities.CreateAGroupOwner()

	blockchain := entities.CreateBlockChain()
	sh, _ := ipfs.InitIPFS()
	proxy := entities.CreateIPFSProxy()

	operator := entities.CreateOperator(proxy, sh, blockchain)

	_, err := groupOwner.ListFiles("123")
	assert.EqualError(t, err, "unable to locate files")

	groupOneUuid := groupOwner.RegisterNewGroup(proxy)

	emptyFiles, err := groupOwner.ListFiles(groupOneUuid)

	assert.Nil(t, err)
	assert.Equal(t, len(emptyFiles), 0)

	groupOneMembers := []entities.GroupMember{}

	goldenFileBytes, err := utils.LoadRawBytesFromFile(TEST_FILEPATH)
	assert.Nil(t, err)

	transactionIDs := []string{}
	IPFSHandles := []string{}
	for i := 0; i < GROUP_ONE_MEMBER_COUNT; i++ {
		groupOneMembers = append(groupOneMembers, entities.CreateAGroupMember())
		_, _, err := groupOneMembers[i].UploadFile(&operator, &groupOwner, groupOneUuid, TEST_FILEPATH)
		assert.EqualError(t, err, "is not a member")

		groupOwner.AddNewMemberObj(proxy, groupOneUuid, groupOneMembers[i])

		assert.Equal(t, groupOneMembers[i].IsMember(), true)
		assert.NotEqual(t, groupOneMembers[i].GetUuid(), "")
		assert.NotNil(t, groupOneMembers[i].GetPublicKey())

		_, _, err = groupOneMembers[i].UploadFile(&operator, &groupOwner, "123", TEST_FILEPATH)
		assert.EqualError(t, err, "is not a member")

		transactionID, IPFSHandle, err := groupOneMembers[i].UploadFile(&operator, &groupOwner, groupOneUuid, TEST_FILEPATH)
		assert.NotEqual(t, transactionID, "")
		assert.NotEqual(t, IPFSHandle, "")
		assert.Nil(t, err)

		files, err := groupOwner.ListFiles(groupOneUuid)
		assert.Nil(t, err)
		assert.Equal(t, len(files), i+1)

		decryptedFilePath, _, err := groupOneMembers[i].DownloadFile(&operator, groupOneUuid, "")
		assert.EqualError(t, err, "transaction ID should not be an empty string")
		os.Remove(decryptedFilePath)

		decryptedFilePath, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, "abc")
		assert.EqualError(t, err, "could not locate transaction")
		os.Remove(decryptedFilePath)

		decryptedFilePath, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, transactionID)
		assert.Nil(t, err)

		decryptedFileRawBytes, err := utils.LoadRawBytesFromFile(decryptedFilePath)
		assert.Nil(t, err)

		assert.Equal(t, goldenFileBytes, decryptedFileRawBytes)

		os.Remove(decryptedFilePath)

		transactionIDs = append(transactionIDs, transactionID)
		IPFSHandles = append(IPFSHandles, IPFSHandle)
	}

	for i := 0; i < GROUP_ONE_MEMBER_COUNT; i++ {
		decryptedFilePath, _, err := groupOneMembers[i].DownloadFile(&operator, groupOneUuid, transactionIDs[i])
		assert.Nil(t, err)

		decryptedFileRawBytes, err := utils.LoadRawBytesFromFile(decryptedFilePath)
		assert.Nil(t, err)

		assert.Equal(t, goldenFileBytes, decryptedFileRawBytes)
		os.Remove(decryptedFilePath)

		groupOwner.RemoveMemberObjAndSecureFiles(&operator, groupOneUuid, groupOneMembers[i])

		decryptedFilePath, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, transactionIDs[i])
		assert.EqualError(t, err, "crypto/rsa: decryption error")
		os.Remove(decryptedFilePath)

		decryptedFilePath, _, err = groupOwner.DownloadFile(&operator, groupOneUuid, transactionIDs[i])
		assert.EqualError(t, err, "crypto/rsa: decryption error")
		os.Remove(decryptedFilePath)

		files, err := groupOwner.ListFiles(groupOneUuid)
		assert.Nil(t, err)

		decryptedFilePath, _, err = groupOwner.DownloadFile(&operator, groupOneUuid, files[i].TransactionID)
		assert.Nil(t, err)

		os.Remove(decryptedFilePath)

		err = ipfs.DownloadFileFromIPFS(sh, IPFSHandles[i], "")
		assert.Nil(t, err)

		unauthorizedIPFSDownload, err := utils.LoadRawBytesFromFile(IPFSHandles[i])
		assert.Nil(t, err)
		assert.NotEqual(t, goldenFileBytes, unauthorizedIPFSDownload)

		os.Remove(IPFSHandles[i])
	}

	files, err := groupOwner.ListFiles(groupOneUuid)
	assert.Nil(t, err)
	for i := 0; i < GROUP_ONE_MEMBER_COUNT; i++ {
		groupOwner.AddNewMemberObj(proxy, groupOneUuid, groupOneMembers[i])

		decryptedFilePath, _, err := groupOwner.DownloadFile(&operator, groupOneUuid, files[i].TransactionID)
		assert.Nil(t, err)

		os.Remove(decryptedFilePath)

		groupOwner.RemoveMemberObj(&operator, groupOneUuid, groupOneMembers[i])
		decryptedFilePath, _, err = groupOwner.DownloadFile(&operator, groupOneUuid, files[i].TransactionID)
		assert.Nil(t, err)

		os.Remove(decryptedFilePath)
	}

	err = cleanup()
	assert.Nil(t, err)
}

func cleanup() error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() {
			os.RemoveAll(entry.Name())
			continue
		}
		if ext != ".go" && ext != ".txt" {
			os.Remove(entry.Name())
		}
	}

	return nil
}

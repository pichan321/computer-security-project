package tests

import (
	"blockchain-fileshare/entities"
	"blockchain-fileshare/ipfs"
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
		assert.Equal(t, len(files), i + 1)

		_, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, "")
		assert.EqualError(t, err, "transaction ID should not be an empty string")

		_, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, "abc")
		assert.EqualError(t, err, "could not locate transaction")

		_, _, err = groupOneMembers[i].DownloadFile(&operator, groupOneUuid, transactionID)
		assert.EqualError(t, err, "could not locate transaction")
	
	}

	for i := 0; i < GROUP_ONE_MEMBER_COUNT; i++ {

	}

}

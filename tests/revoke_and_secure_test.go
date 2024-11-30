package tests

import (
	"blockchain-fileshare/entities"
	"blockchain-fileshare/ipfs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevokeAndSecureFiles(t *testing.T) {
	groupOwner := entities.CreateAGroupOwner()

	blockchain := entities.CreateBlockChain()
	sh, _ := ipfs.InitIPFS()
	proxy := entities.CreateIPFSProxy()

	operator := entities.CreateOperator(proxy, sh, blockchain)

	groupOneMembers := []entities.Member{}

	for i := 0; i < 5; i++ {
		groupOneMembers = append(groupOneMembers, entities.CreateAGroupMember())
	}

	_, err := groupOwner.ListFiles("123")

	groupOneUuid := groupOwner.RegisterNewGroup(proxy)

	emptyFiles, err := groupOwner.ListFiles(groupOneUuid)
	if err


}

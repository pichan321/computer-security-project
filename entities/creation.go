package entities

import (
	"blockchain-fileshare/utils"

	"github.com/google/uuid"
)

func CreateAGroupOwner() GroupOwner {
	uuid := uuid.New().String()[:6]
	public, private := utils.GenerateKeyPair(uuid)
	g := GroupOwner{
		uuid:        uuid,
		groupsOwned: []Group{},
		publicKey:   public,
		privateKey:  private,
	}
	return g
}

func CreateAGroupMember() GroupMember {
	uuid := uuid.New().String()[:6]
	public, private := utils.GenerateKeyPair(uuid)
	g := GroupMember{
		uuid:             uuid,
		groupsAssociated: []Group{},
		publicKey:        public,
		privateKey:       private,
	}
	return g
}

package entities

import (
	keys "blockchain-fileshare/keys"

	"github.com/google/uuid"
)

func CreateAGroupOwner() GroupOwner {
	uuid := uuid.New().String()[:6]
	public, private := keys.GenerateKeyPair(uuid)
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
	public, private := keys.GenerateKeyPair(uuid)
	g := GroupMember{
		uuid:             uuid,
		groupsAssociated: []Group{},
		publicKey:        public,
		privateKey:       private,
	}
	return g
}

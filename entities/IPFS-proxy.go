package entities

type UserMetadata struct {
	uuid string
	publicKey []byte
}

//for the distributed access control policies and group key management (Huang et al.)
type GroupMetadata struct {
	ownerUuid string
	publicKey []byte
	privateKey []byte
	users []UserMetadata
}

type IPFSProxy struct {
	groups map[string]GroupMetadata

}
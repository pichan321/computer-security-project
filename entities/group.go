package entities

import (
	keys "blockchain-fileshare/keys"
	"blockchain-fileshare/utils"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Member interface {
	ReadFile(proxy *IPFSProxy, groupID string, filename string) error
	DownloadFile(proxy *IPFSProxy, groupID string, filename string) error
	UploadFile(proxy *IPFSProxy, groupID string, filepath string) error
	DeleteFile(proxy *IPFSProxy, groupID string, filename string) error
	IsMember() bool
	IsOwner() bool
	GetUuid() string
	GetPublicKey() []byte
}

type File struct {
	fileOwner []Member
	handle    string
}

type Group struct {
	groupID      string
	groupMembers []Member
	files        []File
}

type GroupOwner struct {
	uuid        string
	groupsOwned []Group
	publicKey   []byte
	privateKey  []byte
}

type GroupMember struct {
	uuid             string
	groupsAssociated []Group
	publicKey        []byte
	privateKey       []byte
}

type UserMetadata struct { //this is like GroupMember/GroupOwner but since we don't want private key to be stored in the proxy, I chose to go with this struct
	uuid      string
	publicKey []byte
}

// for the distributed access control policies and group key management (Huang et al.)
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

func (proxy IPFSProxy) PrintUsers(groupID string) {
	for _, m := range proxy.groups[groupID].users {
		fmt.Println(m.uuid)
	}
}

func (g GroupOwner) IsMemberOf(proxy *IPFSProxy, groupID string) bool {
	for _, m := range proxy.groups[groupID].users {
		if m.uuid == g.GetUuid() {
			return true
		}
	}

	return false
}

func (g GroupMember) IsMemberOf(proxy *IPFSProxy, groupID string) bool {
	for _, m := range proxy.groups[groupID].users {
		if m.uuid == g.GetUuid() {
			return true
		}
	}

	return false
}

// this function is necessary because private key of each GroupOwner should not be exposed by any means
func (g GroupOwner) SignSignature(filePath string) ([]byte, error) {
	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// this function is necessary because private key of each GroupMember should not be exposed by any means
func (g GroupMember) SignSignature(filePath string) ([]byte, error) {
	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func CreateIPFSProxy() *IPFSProxy {
	return &IPFSProxy{
		groups: map[string]GroupMetadata{},
	}
}

func (g *GroupOwner) RegisterNewGroup(proxy *IPFSProxy) string {
	groupUuid := uuid.New().String()[:6]
	public, private := keys.GenerateKeyPair(groupUuid)

	// newG := GroupOwner{
	// 	uuid:        g.GetUuid(),
	// 	groupsOwned: g.groupsOwned,
	// 	publicKey:   public,
	// 	privateKey:  private,
	// }
	group := Group{ //this is stored with the group owner
		groupID:      groupUuid,
		groupMembers: []Member{g},
		files:        []File{},
	}

	groupMetadata := GroupMetadata{ //this is stored with IPFS Proxy
		ownerUuid:  g.GetUuid(),
		groupUuid:  groupUuid,
		publicKey:  public,
		privateKey: private,
		users: []UserMetadata{
			UserMetadata{
				uuid:      g.GetUuid(),
				publicKey: g.publicKey,
			},
		},
	}

	g.groupsOwned = append(g.groupsOwned, group)
	(*proxy).groups[groupUuid] = groupMetadata
	return groupUuid
}

func isValidMember(memberUuid string, allUsers []Member) (Member, bool) {
	for _, member := range allUsers {
		if member.GetUuid() == memberUuid {
			return member, true
		}
	}

	return nil, false
}

func (g *GroupOwner) AddNewMember(groupID string, memberUuid string, allUsers []Member) error {
	member, isValid := isValidMember(memberUuid, allUsers)
	if !isValid {
		return errors.New("invalid member/user uuid")
	}

	for _, group := range g.groupsOwned {
		if group.groupID == groupID {
			group.groupMembers = append(group.groupMembers, member)
			return nil
		}
	}

	return errors.New("unexpected error while adding new member to the group")
}

func (g *GroupOwner) registerNewMemberInIPFSProxy(proxy *IPFSProxy, groupUuid string, member Member) error {
	groupMetadata, exists := proxy.groups[groupUuid]
	if !exists {
		return errors.New("group does not exist!")
	}

	for _, m := range groupMetadata.users {
		if m.uuid == member.GetUuid() {
			return errors.New("user was already added!")
		}
	}

	groupMetadata.users = append(groupMetadata.users, UserMetadata{
		uuid:      member.GetUuid(),
		publicKey: member.GetPublicKey(),
	})

	proxy.groups[groupUuid] = groupMetadata
	return nil
}

func (g *GroupOwner) removeMemberInIPFSProxy(proxy *IPFSProxy, groupUuid string, member Member) error {
	groupMetadata, exists := proxy.groups[groupUuid]
	if !exists {
		return errors.New("group does not exist!")
	}

	i := -1
	for idx, m := range groupMetadata.users {
		if m.uuid == member.GetUuid() {
			i = idx
			break
		}
	}
	if i == -1 {
		return errors.New("user not found!")
	}

	groupMetadata.users = append(groupMetadata.users[:i], groupMetadata.users[i+1:]...)
	proxy.groups[groupUuid] = groupMetadata
	return nil
}

func (g *GroupOwner) AddNewMemberObj(proxy *IPFSProxy, groupID string, member Member) error {
	fmt.Println("group to find", groupID)
	fmt.Println(g.groupsOwned)
	for _, group := range g.groupsOwned {

		if group.groupID == groupID {
			group.groupMembers = append(group.groupMembers, member)
			g.registerNewMemberInIPFSProxy(proxy, groupID, member)
			return nil
		}
	}

	return errors.New("unexpected error while adding new member to the group")
}

func (g *GroupOwner) RemoveMemberObj(proxy *IPFSProxy, groupID string, member Member) error {
	memberToBeRemoved := GroupMember{}
	groupsOwned := g.groupsOwned
	gIndex := -1
	mIndex := -1
	fmt.Println("Member id", member.GetUuid())
	for gIdx, group := range groupsOwned {
		if group.groupID == groupID {
			gIndex = gIdx
			members := group.groupMembers
			for idx, m := range members {
				if m.GetUuid() == member.GetUuid() {
					mIndex = idx
					g.removeMemberInIPFSProxy(proxy, groupID, memberToBeRemoved)
					break
				}
			}

			break
		}
	}

	if gIndex == -1 || mIndex == -1 {
		fmt.Println("Group INdex", gIndex, "M index", mIndex)
		return errors.New("unexpected error while removing member from the group")
	}

	g.groupsOwned[gIndex].groupMembers = append(g.groupsOwned[gIndex].groupMembers[:mIndex], g.groupsOwned[gIndex].groupMembers[mIndex+1:]...)

	return errors.New("unexpected error while removing member from the group")
}

func (g *GroupOwner) RemoveMember(groupID string, memberUuid string, allUsers []Member) error {
	member, isValid := isValidMember(memberUuid, allUsers)
	if !isValid {
		return errors.New("invalid member/user UUID")
	}

	for i, group := range g.groupsOwned {
		if group.groupID == groupID {

			memberIndex := -1
			for j, m := range group.groupMembers {
				if m.GetUuid() == member.GetUuid() {
					memberIndex = j
					break
				}
			}

			if memberIndex == -1 {
				return errors.New("member not found in the specified group")
			}

			group.groupMembers = append(group.groupMembers[:memberIndex], group.groupMembers[memberIndex+1:]...)
			g.groupsOwned[i] = group

			return nil
		}
	}
	return errors.New("group not found")
}

// ReadFile(groupID string, filename string)
// DownloadFile(groupID string, filename string)
// UploadFile(groupID string, filepath string)
// DeleteFile(groupID string, filename string)

func (g GroupOwner) ReadFile(proxy *IPFSProxy, groupID string, filename string) error {
	return nil

}

func (g GroupOwner) DownloadFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupOwner) UploadFile(proxy *IPFSProxy, groupID string, filepath string) error {

	return nil
}

func (g GroupOwner) DeleteFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupOwner) IsMember() bool {

	return false
}

func (g GroupOwner) IsOwner() bool {
	return true
}

func (g GroupOwner) GetUuid() string {
	return g.uuid
}

func (g GroupOwner) GetPublicKey() []byte {
	return g.publicKey
}

func (g GroupMember) ReadFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupMember) DownloadFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupMember) UploadFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupMember) DeleteFile(proxy *IPFSProxy, groupID string, filename string) error {

	return nil
}

func (g GroupMember) IsMember() bool {

	return false
}

func (g GroupMember) IsOwner() bool {

	return false
}

func (g GroupMember) GetUuid() string {
	return g.uuid
}

func (g GroupMember) GetPublicKey() []byte {
	return g.publicKey
}

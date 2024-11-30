package entities

import (
	keys "blockchain-fileshare/keys"
	"blockchain-fileshare/utils"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

type GroupOwner struct {
	uuid        string
	groupsOwned []Group
	publicKey   []byte
	privateKey  []byte
}

func (g GroupOwner) IsMemberOf(proxy *IPFSProxy, groupID string) (bool, error) {
	for _, m := range proxy.groups[groupID].users {
		if m.uuid == g.GetUuid() {
			return true, nil
		}
	}

	return false, errors.New("is not a member")
}

// this function is necessary because private key of each GroupOwner should not be exposed by any means
func (g GroupOwner) SignSignature(filePath string) ([]byte, error) {
	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (g *GroupOwner) RegisterNewGroup(proxy *IPFSProxy) string {
	groupUuid := uuid.New().String()[:6]
	public, private := keys.GenerateKeyPair(groupUuid)

	newG := GroupOwner{
		uuid:        g.GetUuid(),
		groupsOwned: g.groupsOwned,
		publicKey:   public,
		privateKey:  private,
	}
	group := Group{ //this is stored with the group owner
		groupID:      groupUuid,
		groupMembers: []Member{newG},
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

func (g GroupOwner) ListFiles(groupID string) ([]File, error) {
	for _, group := range g.groupsOwned {
		if group.groupID == groupID {
			return group.files, nil
		}
	}
	return nil, errors.New("unable to locate files")
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
	for idx, group := range g.groupsOwned {

		if group.groupID == groupID {
			group.groupMembers = append(group.groupMembers, member)
			g.groupsOwned[idx].groupMembers = group.groupMembers
			g.registerNewMemberInIPFSProxy(proxy, groupID, member)
			return nil
		}
	}

	return errors.New("unexpected error while adding new member to the group")
}

func (g *GroupOwner) RemoveMemberObj(operator *Operators, groupID string, member Member) error {
	memberToBeRemoved := GroupMember{}
	groupsOwned := g.groupsOwned
	gIndex := -1
	mIndex := -1
	fmt.Println("Member id", member.GetUuid())
	for gIdx, group := range groupsOwned {
		if group.groupID == groupID {
			gIndex = gIdx
			members := group.groupMembers
			fmt.Println("memers", members)
			for idx, m := range members {

				if m.GetUuid() == member.GetUuid() {

					mIndex = idx
					break
				}
			}

			break
		}
	}

	fmt.Println("group index", gIndex, "m index", mIndex)
	if gIndex == -1 || mIndex == -1 {
		fmt.Println("Group INdex", gIndex, "M index", mIndex)
		return errors.New("unexpected error while removing member from the group")
	}

	g.removeMemberInIPFSProxy(operator.proxy, groupID, memberToBeRemoved)
	g.groupsOwned[gIndex].groupMembers = append(g.groupsOwned[gIndex].groupMembers[:mIndex], g.groupsOwned[gIndex].groupMembers[mIndex+1:]...)
	return nil
}

func (g *GroupOwner) RemoveMemberObjAndSecureFiles(operator *Operators, groupID string, member Member) error {
	memberToBeRemoved := GroupMember{}
	groupsOwned := g.groupsOwned
	gIndex := -1
	mIndex := -1
	fmt.Println("Member id", member.GetUuid())
	for gIdx, group := range groupsOwned {
		if group.groupID == groupID {
			gIndex = gIdx
			members := group.groupMembers
			fmt.Println("memers", members)
			for idx, m := range members {

				if m.GetUuid() == member.GetUuid() {

					mIndex = idx
					break
				}
			}

			break
		}
	}

	fmt.Println("group index", gIndex, "m index", mIndex)
	if gIndex == -1 || mIndex == -1 {
		fmt.Println("Group INdex", gIndex, "M index", mIndex)
		return errors.New("unexpected error while removing member from the group")
	}

	g.removeMemberInIPFSProxy(operator.proxy, groupID, memberToBeRemoved)
	g.groupsOwned[gIndex].groupMembers = append(g.groupsOwned[gIndex].groupMembers[:mIndex], g.groupsOwned[gIndex].groupMembers[mIndex+1:]...)
	operator.proxy.ChangeKeyAndSecureFiles(operator, g, gIndex, groupID) //this is the most crucial part for our threat model
	return nil
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

func (g GroupOwner) ReadFile(operator *Operators, groupID string, filename string) error {
	return nil
}

/*
*
In real world, when a member of the group wants to access a file, they might see it by some abitrary filename.
When they do click the file they intend to download, only transactionHash is used in the process of file retrieval.
*
*/
func (g GroupOwner) DownloadFile(operator *Operators, groupID string, transactionHash string) (string, string, error) {
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

	file, encryptedGroupPrivateKey, err := operator.proxy.DownloadFileFromIPFS(operator.sh, downloadRequest)
	if err != nil {
		return "", "", err
	}

	decryptedGroupPrivateKey, err := utils.DecryptKey(encryptedGroupPrivateKey, g.privateKey)
	
	decryptedFilePath, checksumHash, err := utils.DecryptFile(file, decryptedGroupPrivateKey)
	if err != nil {
		return "", "", err
	}

	return decryptedFilePath, checksumHash, nil
}

func (g *GroupOwner) UploadFile(operator *Operators, groupID string, filePath string) (string, error) {
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
	for idx, group := range g.groupsOwned {
		if group.groupID == groupID {
			groupIdx = idx
			break
		}
	}
	if groupIdx == -1 {
		return "", errors.New("unexpected error while finding group to insert the uploaded file metadata into")
	}

	g.groupsOwned[groupIdx].files = append(g.groupsOwned[groupIdx].files, file)
	return transactionHash, nil
}

func (g GroupOwner) DeleteFile(operator *Operators, groupID string, filename string) error {
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

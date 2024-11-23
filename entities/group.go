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
	"path/filepath"

	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
)

// this struct is to make life easy to deal with Member interface change
type Operators struct {
	proxy      *IPFSProxy
	sh         *shell.Shell
	blockchain *Blockchain
}

type Member interface {
	ReadFile(operator *Operators, groupID string, filename string) error
	DownloadFile(operator *Operators, groupID string, filename string) error
	UploadFile(operator *Operators, groupID string, filepath string) (string, error)
	DeleteFile(operator *Operators, groupID string, filename string) error
	IsMember() bool
	IsOwner() bool
	GetUuid() string
	GetPublicKey() []byte
}

type File struct {
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
		return nil, errors.New("invalid public key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.New("error parsing public key")
	}

	checksum := sha256.New()
	checksum.Write(downloadRequestBytes)
	hash := checksum.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, err
	}

	return signature, nil
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

	publicKey, err := x509.ParsePKCS8PrivateKey(publicKeyBlock.Bytes)
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

	err = rsa.VerifyPSS(publicKey.(*rsa.PublicKey), crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (proxy IPFSProxy) DownloadFileFromIPFS(sh *shell.Shell, downloadRequest DownloadRequest) (error) {
	err := ipfs.DownloadFileFromIPFS(sh, downloadRequest.IPFSHandle)
	if err != nil {
		return err
	}

	// encryptedFileName := downloadRequest.IPFSHandle
	return nil


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

func (g GroupOwner) IsMemberOf(proxy *IPFSProxy, groupID string) (bool, error) {
	for _, m := range proxy.groups[groupID].users {
		if m.uuid == g.GetUuid() {
			return true, nil
		}
	}

	return false, errors.New("is not a member")
}

func (g GroupMember) IsMemberOf(proxy *IPFSProxy, groupID string) (bool, error) {
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

func (g GroupOwner) ReadFile(operator *Operators, groupID string, filename string) error {
	return nil
}

/*
*
In real world, when a member of the group wants to access a file, they might see it by some abitrary filename.
When they do click the file they intend to download, only transactionHash is used in the process of retrieval.
*
*/
func (g GroupOwner) DownloadFile(operator *Operators, groupID string, transactionHash string) error {
	data, err := operator.blockchain.GetTransactionByHash(transactionHash)
	if err != nil {
		return nil
	}

	downloadRequest := DownloadRequest{
		requestedUserId: g.GetUuid(),
		groupId:         groupID,
		IPFSHandle:      data.IPFSHash,
	}

	signature, err := SignDownloadRequest(downloadRequest, g.privateKey)
	if err != nil {
		return err
	}

	_, err = operator.proxy.VerifyDownloadReqSignature(downloadRequest, signature)
	if err != nil {
		return err
	}

	file, groupPrivateKey, err := operator.proxy.DownloadFile(downloadRequest)

	return nil
}

func (g GroupOwner) UploadFile(operator *Operators, groupID string, filePath string) (string, error) {
	if isMember, err := g.IsMemberOf(operator.proxy, groupID); !isMember {
		return "", err
	}

	signature, err := utils.SignSignature(filePath, g.privateKey)
	if err != nil {
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
		userId:   g.GetUuid(),
		groupId:  groupID,
		fileHash: checksum,
		IPFSHash: handle,
	}
	transactionHash := operator.blockchain.CreateTransaction(transactionData)

	file := File{
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

func (g GroupMember) ReadFile(operator *Operators, groupID string, filename string) error {

	return nil
}

func (g GroupMember) DownloadFile(operator *Operators, groupID string, filename string) error {

	return nil
}

func (g GroupMember) UploadFile(operator *Operators, groupID string, filename string) (string, error) {

	return "", nil
}

func (g GroupMember) DeleteFile(operator *Operators, groupID string, filename string) error {

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

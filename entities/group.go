package entities

import (
	"blockchain-fileshare/utils"
	"errors"

	"github.com/google/uuid"
)

type Member interface {
	ReadFile(groupID string, filename string)
	DownloadFile(groupID string, filename string)
	UploadFile(groupID string, filepath string)
	DeleteFile(groupID string, filename string)
	IsMember() bool
	IsOwner() bool
	GetUuid() string
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
}

type GroupMember struct {
	uuid             string
	groupsAssociated []Group
}

func (g GroupOwner) RegisterNewGroup() string {
	groupUuid := uuid.New().String()
	utils.GenerateKeyPair(groupUuid)
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

func (g GroupOwner) ReadFile(groupID string, filename string) {

}

func (g GroupOwner) DownloadFile(groupID string, filename string) {

}

func (g GroupOwner) UploadFile(groupID string, filepath string) {

}

func (g GroupOwner) DeleteFile(groupID string, filename string) {

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

func (g GroupMember) ReadFile(groupID string, filename string) {

}

func (g GroupMember) DownloadFile(groupID string, filename string) {

}

func (g GroupMember) UploadFile(groupID string, filepath string) {

}

func (g GroupMember) DeleteFile(groupID string, filename string) {

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
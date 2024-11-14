package entities

type Member interface {
	ReadFile(groupID string, filename string)
	DownloadFile(groupID string, filename string)
	UploadFile(groupID string, filepath string)
	DeleteFile(groupID string, filename string)
	IsMember() bool
	IsOwner() bool
}

type File struct {
	fileOwner []Member
	handle string
}

type Group struct {
	groupID      string
	groupMembers []Member
	files []File
}

type GroupOwner struct {
	uuid        string
	groupsOwned []Group
}

type GroupMember struct {
	uuid             string
	groupsAssociated []Group
}

func (g GroupOwner) RegisterNewGroup() {

}

func (g GroupOwner) AddNewMember(groupID string, memberUuid string) {

}

func (g GroupOwner) RemoveMember(groupID string, memberUuid string) {

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

	return false
}

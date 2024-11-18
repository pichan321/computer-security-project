package main

import (
	"blockchain-fileshare/entities"
)

func main() {

	// utils.GenerateKeyPair("user1")
	// utils.InitIPFS()
	// f, _ := os.Open("user1_public_key.pem")
	// b, _ := io.ReadAll(f)

	// err = utils.DecryptFile("sample.txt-encrypted.txt", b)
	// fmt.Println(err)
	proxy := entities.CreateIPFSProxy()
	// sh, _ := utils.InitIPFS()

	// handle, err := utils.UploadFileToIPFS(sh, "sample.txt", b)

	// o, _ := os.Open("user1_private_key.pem")
	// b, _ = io.ReadAll(o)
	// fmt.Println(handle)
	// utils.DownloadFileFromIPFS(sh, proxy, handle)
	// utils.DecryptFile(handle, b)

	groupOwner := entities.CreateAGroupOwner()
	groupUuid := groupOwner.RegisterNewGroup(proxy)

	member := entities.CreateAGroupMember()

}

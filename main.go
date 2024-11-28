package main

import (
	"blockchain-fileshare/entities"
	"blockchain-fileshare/ipfs"
	"fmt"
)

func main() {

	// utils.GenerateKeyPair("user1")
	// utils.InitIPFS()
	// f, _ := os.Open("user1_public_key.pem")
	// b, _ := io.ReadAll(f)

	// err = utils.DecryptFile("sample.txt-encrypted.txt", b)
	// fmt.Println(err)

	blockchain := entities.CreateBlockChain()
	sh, _ := ipfs.InitIPFS()
	proxy := entities.CreateIPFSProxy()
	// handle, _, err := ipfs.UploadFileToIPFS(sh, "sample.txt", b)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	operator := entities.CreateOperator(proxy, sh, blockchain)
	// o, _ := os.Open("user1_private_key.pem")
	// b, _ = io.ReadAll(o)
	// fmt.Println(handle)
	// ipfs.DownloadFileFromIPFS(sh, handle)
	// utils.DecryptFile(handle, b)

	// signature, err := utils.SignSignature("sample.txt", b)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(signature)

	groupOwner := entities.CreateAGroupOwner()
	groupUuid := groupOwner.RegisterNewGroup(proxy)

	member1 := entities.CreateAGroupMember()
	// proxy.PrintUsers(groupUuid)
	// proxy.PrintUsers(groupUuid)
	// fmt.Println()
	// err := groupOwner.AddNewMemberObj(proxy, groupUuid, member1)
	// fmt.Println(err)
	// proxy.PrintUsers(groupUuid)
	// fmt.Println()
	// err = groupOwner.RemoveMemberObj(proxy, groupUuid, member1)
	// fmt.Println(err)
	// proxy.PrintUsers(groupUuid)

	transID, err := groupOwner.UploadFile(&operator, groupUuid, "sample.txt")
	if err != nil {
		fmt.Println(err)
	}

	groupOwner.AddNewMemberObj(proxy, groupUuid, member1)

	// err = groupOwner.DownloadFile(&operator, groupUuid, transID)
	// fmt.Println("Error", err)
	err = member1.DownloadFile(&operator, groupUuid, transID)
	fmt.Println("Error", err)
}

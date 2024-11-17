package main

import (
	"blockchain-fileshare/utils"
	"fmt"
	"io"
	"os"
)

func main() {

	// utils.GenerateKeyPair("user1")
	// utils.InitIPFS()
	f, _ := os.Open("user1_public_key.pem")
	b, _ := io.ReadAll(f)
	fmt.Println(string(b))
	err := utils.EncryptFile("sample.txt", b)
	fmt.Println(err)

	o, _ := os.Open("user1_private_key.pem")
	b, _ = io.ReadAll(o)

	err = utils.DecryptFile("sample.txt-encrypted.txt", b)
	fmt.Println(err)
}

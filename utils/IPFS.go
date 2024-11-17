package utils

import (
	"fmt"
	"os"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

const IPFS_ENDPOINT = `https://ipfs.io/ipfs/`

func InitIPFS() {
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
	}

	fmt.Println(cid)
}

func UploadFileToIPFS() {

}

func DeleteFileFromIPFS() {
	
}

package utils

import (
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

const IPFS_ENDPOINT = `https://ipfs.io/ipfs/`

func InitIPFS() (*shell.Shell, error) {
	sh := shell.NewShell("localhost:5001")
	_, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		return nil, err
	}

	return sh, nil
}

func UploadFileToIPFS(sh *shell.Shell, filePath string) {

}

func DeleteFileFromIPFS(sh *shell.Shell, filePath string) {

}

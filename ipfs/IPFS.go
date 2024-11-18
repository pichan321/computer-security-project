package ipfs

import (
	"blockchain-fileshare/entities"
	"blockchain-fileshare/utils"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

const IPFS_ENDPOINT = `https://ipfs.io/ipfs/`

func InitIPFS() (*shell.Shell, error) {
	sh := shell.NewShell("localhost:5001")
	h, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		return nil, err
	}

	fmt.Println(h)
	return sh, nil
}

func UploadFileToIPFS(sh *shell.Shell, proxy *entities.IPFSProxy, filePath string, groupPublicKeyBytes []byte) (string, error) {

	

	filename, err := utils.EncryptFile(filePath, groupPublicKeyBytes)
	if err != nil {
		return "", err
	}

	f, err := os.Open(filename)
	if err != nil {
		return "", errors.New("here")
	}
	defer f.Close()
	defer os.Remove(filename)

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	hash, err := sh.Add(bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	return hash, nil
}

func DownloadFileFromIPFS(sh *shell.Shell, proxy *entities.IPFSProxy, handle string) {
	sh.Get(handle, "./")
}

func DeleteFileFromIPFS(sh *shell.Shell, fileHandle string) {

}

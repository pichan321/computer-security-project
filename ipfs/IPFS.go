package ipfs

import (
	"blockchain-fileshare/utils"
	"bytes"
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

func UploadFileToIPFS(sh *shell.Shell, filePath string, publicKeyBytes []byte) (string, string, error) {
	filename, checksum, err := utils.EncryptFile(filePath, publicKeyBytes)
	if err != nil {
		return "", "", err
	}

	f, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}

	defer f.Close()
	defer os.Remove(filename)

	b, err := io.ReadAll(f)
	if err != nil {
		return "", "", err
	}

	hash, err := sh.Add(bytes.NewReader(b))
	if err != nil {
		return "", "", err
	}

	return hash, checksum, nil
}

func DownloadFileFromIPFS(sh *shell.Shell, handle string, fileExtension string) error {
	err := sh.Get(handle, fmt.Sprintf(`%s%s`, handle, fileExtension))
	return err
}

func DeleteFileFromIPFS(sh *shell.Shell, handle string) error {
	return sh.Unpin(handle)
}

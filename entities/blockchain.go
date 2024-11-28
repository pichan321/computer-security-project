package entities

import (
	"errors"

	"github.com/google/uuid"
)

type Data struct {
	userId   string
	groupId  string
	fileHash string
	IPFSHash string
	fileExtension string
}

type Blockchain struct {
	blocks map[string]Data
}

func (b *Blockchain) CreateTransaction(data Data) string {
	uuid := uuid.New().String()
	b.blocks[uuid] = data
	return uuid
}

func (b Blockchain) GetTransactionByHash(transactionId string) (Data, error) {
	data, ok := b.blocks[transactionId]
	if !ok {
		return Data{}, errors.New("could not locate transaction")
	}

	return data, nil
}

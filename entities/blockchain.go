package entities

import "github.com/google/uuid"

type Data struct {
	groupId  string
	fileHash string
	IPFSHash string
}

type Blockchain struct {
	blocks map[string]Data
}

func (b *Blockchain) CreateTransaction(data Data) string {
	uuid := uuid.New().String()
	b.blocks[uuid] = data
	return uuid
}

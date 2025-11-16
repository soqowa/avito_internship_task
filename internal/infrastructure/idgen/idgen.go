package idgen

import "github.com/google/uuid"

type IDGenerator interface {
	Generate() string
}

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

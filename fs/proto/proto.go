package proto

//go:generate gencode go -package proto -schema proto.schema

const (
	ItemFile = iota
	ItemDir
)
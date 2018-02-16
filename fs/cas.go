package fs

type CAS interface {
	Read(hash []byte) ([]byte, error)
	Write(data []byte) ([]byte, error)
}

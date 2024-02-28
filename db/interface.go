package db

type Db interface {
	Get(int64) ([]byte, bool)
	Set([]byte, int64)
	GetLeafLength() int64
	SetLeafLength(int64)
	Serialize() ([]byte, error)
}

var _ Db = &Memorybaseddb{}
var _ Db = &Filebaseddb{}

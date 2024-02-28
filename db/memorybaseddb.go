package db

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

type Memorybaseddb struct {
	leafLength int64
	nodes      map[int64][]byte
}

func NewMemorybaseddb(leafLength int64, nodes map[int64][]byte) *Memorybaseddb {
	db := Memorybaseddb{nodes: nodes, leafLength: leafLength}
	return &db
}
func (db *Memorybaseddb) Get(index int64) ([]byte, bool) {
	value, ok := db.nodes[index]
	return value, ok
}
func (db *Memorybaseddb) Set(value []byte, index int64) {
	//require correct wordsize
	db.nodes[index] = value
}
func (db *Memorybaseddb) GetLeafLength() int64 {
	return db.leafLength
}
func (db *Memorybaseddb) SetLeafLength(value int64) {
	db.leafLength = value
}

type Thing struct {
	length [][]byte
	nodes  [][][]byte
}

func FromSerialized(input []byte) (*Memorybaseddb, error) { //likely a much easier way to do this
	var output []any
	err := rlp.DecodeBytes(input, &output)
	if err != nil {
		return nil, fmt.Errorf("Problem rlp decoding db node and length data: %w", err)
	}

	if len(output) != 2 {
		return nil, errors.New("RLP formatting error - Outer tuple should be `[leafLength,nodes]`")
	}

	leafLengthArr, leafLengthOk := output[0].([]byte)
	if !leafLengthOk {
		return nil, errors.New("RLP formatting error - `leafLength` unable to cast to `[]byte`")
	}

	leafLengthBuffer := make([]byte, 8-len(leafLengthArr))
	leafLengthBuffer = append(leafLengthBuffer, leafLengthArr...)
	leafLength := int64(binary.BigEndian.Uint64(leafLengthBuffer))
	nodesArr, nodesArrayOk := output[1].([]any)

	if !nodesArrayOk {
		return nil, errors.New("RLP formatting error - `nodesArray` unable to cast to `[][][]byte`")
	}

	nodes := make(map[int64][]byte, len(nodesArr))

	for _, _keyPair := range nodesArr {
		keyPair, ok := _keyPair.([]any)
		if !ok {
			return nil, fmt.Errorf("RLP formatting error - `keyPair` unable to cast to `[]any` got %T", _keyPair)
		}
		unpaddedKeyBuffer := keyPair[0].([]byte)
		keyBuffer := make([]byte, 8-len(unpaddedKeyBuffer))
		keyBuffer = append(keyBuffer, unpaddedKeyBuffer...)
		k := int64(binary.BigEndian.Uint64(keyBuffer))
		nodes[k] = keyPair[1].([]byte)
	}

	db := Memorybaseddb{nodes: nodes, leafLength: leafLength}
	return &db, nil
}

func (db *Memorybaseddb) Serialize() ([]byte, error) {
	output := []interface{}{}

	leafLengthBytes, err := rlp.EncodeToBytes(uint(db.GetLeafLength()))
	if err != nil {
		return nil, errors.New("problem representing leafLength in []byte")
	}

	nodes := [][][]byte{}
	for nodeIndex, nodeValue := range db.nodes {
		k, _ := rlp.EncodeToBytes(uint(nodeIndex))
		kv := [][]byte{}
		kv = append(kv, k)
		kv = append(kv, nodeValue)
		nodes = append(nodes, kv)
	}

	output = append(output, leafLengthBytes)
	output = append(output, nodes)

	serializedOutput, err := rlp.EncodeToBytes(output)
	if err != nil {
		return nil, errors.New("Problem rlp encoding db (nodes and leafLength)")
	}

	return serializedOutput, nil
}

package utils

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"time"
)

type HashWithSeed struct {
	Seed []byte
	hash.Hash32
}

func NewHashWithSeed(seed []byte) HashWithSeed {
	h := fnv.New32a()
	return HashWithSeed{
		Seed:   seed,
		Hash32: h,
	}
}

func (h HashWithSeed) Write(data []byte) (int, error) {
	return h.Hash32.Write(data)
}

func (h HashWithSeed) Sum(b []byte) []byte {
	return h.Hash32.Sum(b)
}

func (h HashWithSeed) Reset() {
	h.Hash32.Reset()
}

func (h HashWithSeed) Hash(data string) uint32 {
	h.Reset()
	h.Write(append([]byte(data), h.Seed...))
	return h.Sum32()
}

func CreateHashFunctions(k uint) ([]HashWithSeed, uint) {
	h := make([]HashWithSeed, k)
	ts := uint(time.Now().Unix())
	for i := uint(0); i < k; i++ {
		seed := make([]byte, 32)
		binary.BigEndian.PutUint32(seed, uint32(ts+i))
		hfn := NewHashWithSeed(seed)
		h[i] = hfn
	}
	return h, ts
}

func СreateHashFunctionsWithTS(k, ts uint) ([]HashWithSeed, uint) {
	h := make([]HashWithSeed, k)
	if ts == 0 {
		ts = uint(time.Now().Unix())
	}
	for i := uint(0); i < k; i++ {
		seed := make([]byte, 32)
		binary.BigEndian.PutUint32(seed, uint32(ts+i))
		hfn := NewHashWithSeed(seed)
		h[i] = hfn
	}
	return h, ts
}

func main() { // example of usage
	hashFunctions, _ := CreateHashFunctions(10)
	data := "test-string"

	for _, hfn := range hashFunctions {
		fmt.Println(hfn.Hash(data)) // import fmt**
	}
}

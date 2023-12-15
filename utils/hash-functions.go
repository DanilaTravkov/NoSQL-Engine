package utils

import (
	"encoding/binary"
	"golang.org/x/crypto/blake2b"
	"time"
)

type HashWithSeed struct {
	Seed []byte
}

func (h HashWithSeed) Hash(data []byte) uint64 {
	hash, _ := blake2b.New512(h.Seed)
	//hash.Write(append(data, h.Seed...)) ne treba ako unosimo Seed gore
	hash.Write(data)
	sum := hash.Sum(nil)
	return binary.BigEndian.Uint64(sum)
}

func CreateHashFunctions(k uint) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint(time.Now().Unix())
	for i := uint(0); i < k; i++ {
		seed := make([]byte, 32)
		binary.BigEndian.PutUint32(seed, uint32(ts+i))
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}

//func main() { // example of usage
//	hashFunctions := CreateHashFunctions(10)
//	data := []byte("test-string")
//
//	for _, hfn := range hashFunctions {
//		fmt.Println(hfn.Hash(data)) // import fmt**
//	}
//}

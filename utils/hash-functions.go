package utils

import (
	"encoding/binary"
	"hash/fnv"
	"time"
)

type HashWithSeed struct {
	Seed []byte
}

func (h HashWithSeed) Hash(data string) uint32 {
	hash := fnv.New32a()
	hash.Write(append([]byte(data), h.Seed...))
	sum := hash.Sum(nil)
	return binary.BigEndian.Uint32(sum)
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
//	data := "test-string"
//
//	for _, hfn := range hashFunctions {
//		fmt.Println(hfn.Hash(data)) // import fmt**
//	}
//}

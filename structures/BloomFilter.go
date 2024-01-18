package structures

import (
	"encoding/gob"
	"math"
	"os"
	"projectNASP/utils"
	"strconv"
)

type BloomFilter struct {
	m             uint
	k             uint
	hashFunctions []utils.HashWithSeed
	BitSet        []int
	ts            uint
}

func CreateBloomFilter(expectedElements int, falsePositiveRate float64) BloomFilter {
	b := BloomFilter{}

	b.m = calculateM(expectedElements, falsePositiveRate)
	b.k = calculateK(expectedElements, b.m)
	b.hashFunctions, b.ts = utils.CreateHashFunctions(b.k)
	b.createBitSet()

	return b
}

func (b *BloomFilter) AddElement(element string) {

	for j := 0; j < len(b.hashFunctions); j++ {
		b.hashFunctions[j].Reset()
		b.hashFunctions[j].Write([]byte(element))
		i := b.hashFunctions[j].Sum32() % uint32(b.m)
		b.hashFunctions[j].Reset()
		b.BitSet[i] = 1
	}

}

func (b *BloomFilter) createBitSet() {
	b.BitSet = make([]int, b.m, b.m)
}

func (b *BloomFilter) IsElementInBloomFilter(element string) bool {
	for j := 0; j < len(b.hashFunctions); j++ {
		b.hashFunctions[j].Reset()
		b.hashFunctions[j].Write([]byte(element))
		i := b.hashFunctions[j].Sum32() % uint32(b.m)
		b.hashFunctions[j].Reset()
		if b.BitSet[i] == 0 {
			return false
		}
	}
	return true
}

func (b *BloomFilter) SerializeBloomFilter(gen, lvl int) {
	file, err := os.Create("data/ds/bf/usertable-lvl=" + strconv.Itoa(lvl) + "-gen=" + strconv.Itoa(gen) + "-filter.db")

	if err != nil {
		panic(err)
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(b)

	if err != nil {
		panic(err)
	}

	file.Close()
}

func DeserializeBloomFilter(gen, lvl int) BloomFilter {
	file, err := os.OpenFile("data/ds/bf/usertable-lvl="+strconv.Itoa(lvl)+"-gen="+strconv.Itoa(gen)+"-filter.db", os.O_RDWR, 0777)

	if err != nil {
		panic(err)
	}

	newB := BloomFilter{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&newB)

	if err != nil {
		panic(err)
	}

	newB.hashFunctions, _ = utils.СreateHashFunctionsWithTS(newB.k, newB.ts)
	err = file.Close()

	if err != nil {
		panic(err)
	}

	return newB
}

func calculateM(expectedElements int, falsePositiveRate float64) uint {
	return uint(math.Ceil(float64(expectedElements) * math.Abs(math.Log(falsePositiveRate)) / math.Pow(math.Log(2), float64(2))))
}

func calculateK(expectedElements int, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(expectedElements)) * math.Log(2)))
}

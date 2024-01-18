package structures

import (
	"encoding/gob"
	"math"
	"math/bits"
	"os"
	"projectNASP/utils"
)

const (
	HLL_MIN_PRECISION = 4
	HLL_MAX_PRECISION = 16
)

type HLL struct {
	p     uint8
	m     uint32
	reg   []uint8
	ts    uint
	hashF utils.HashWithSeed
}

func CreateHLL(p uint8) *HLL {
	hll := HLL{p: p}

	hll.m, hll.reg = createBuckets(hll.p)
	hash, ts := utils.CreateHashFunctions(1)
	hll.hashF = hash[0]
	hll.ts = ts

	return &hll
}

func (hll *HLL) Add(key string) {
	hll.hashF.Reset()
	hll.hashF.Write([]byte(key))
	i := hll.hashF.Sum32()
	n := bits.TrailingZeros32(i)
	i = i >> (32 - hll.p)

	hll.reg[i] = uint8(n)

}

func createBuckets(p uint8) (uint32, []uint8) {
	m := uint32(math.Pow(2, float64(p)))
	reg := make([]uint8, m)
	return m, reg
}

func (hll *HLL) emptyCount() int {
	sum := 0
	for _, val := range hll.reg {
		if val == 0 {
			sum++
		}
	}
	return sum
}

func (hll *HLL) Estimate() float64 {
	sum := 0.0
	for _, val := range hll.reg {
		sum = sum + math.Pow(float64(-val), 2.0)
	}

	alpha := 0.7213 / (1.0 + 1.079/float64(hll.m))
	estimation := alpha * math.Pow(float64(hll.m), 2.0) / sum
	emptyRegs := hll.emptyCount()
	if estimation < 2.5*float64(hll.m) {
		if emptyRegs > 0 {
			estimation = float64(hll.m) * math.Log(float64(hll.m)/float64(emptyRegs))
		}
	} else if estimation > math.Pow(2.0, 32.0)/30.0 {
		estimation = -math.Pow(2.0, 32.0) * math.Log(1.0-estimation/math.Pow(2.0, 32.0))
	}
	return estimation
}
func (hll HLL) SerializeHyperLogLog(name string) string {
	name = "data/ds/hll/usertable-" + name + "-HLL.db"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		panic(err)
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(hll)

	if err != nil {
		panic(err)
	}

	file.Close()
	return name
}

func DeserializeHyperLogLog(name string) *HLL {
	file, err := os.OpenFile(name, os.O_RDWR, 0777)

	if err != nil {
		panic(err)
	}

	hll := HLL{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&hll)

	if err != nil {
		panic(err)
	}

	hash, _ := utils.CreateHashFunctions(1)
	hll.hashF = hash[0]

	file.Close()
	return &hll
}

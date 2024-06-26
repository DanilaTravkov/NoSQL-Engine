package structures

import (
	"encoding/gob"
	"math"
	"os"
	"projectDVMVRV/utils"
)

type CountMinSketch struct {
	m             uint // Length of hash-functions
	k             uint // Amount of hash-functions
	hashFunctions []utils.HashWithSeed
	T             [][]uint // Our table with 0-s in the beginning, +1 in each hash-cell if we add an element
	ts            uint
}

func CreateCountMinSketch(e float64, d float64) *CountMinSketch {
	cms := CountMinSketch{}

	cms.m = CalculateM(e)
	cms.k = CalculateK(d)
	cms.hashFunctions, cms.ts = utils.CreateHashFunctions(cms.k)
	cms.T = setT(cms.k, cms.m)

	return &cms
}

func setT(k uint, m uint) [][]uint {
	T := make([][]uint, k)
	for i := range T {
		T[i] = make([]uint, m) // In the beginning 0-s
	}
	return T
}

func (cms *CountMinSketch) Addiction(element string) {

	for i := 0; i < int(cms.k); i++ {
		cms.hashFunctions[i].Reset()
		cms.hashFunctions[i].Write([]byte(element))
		j := cms.hashFunctions[i].Sum32() % uint32(cms.m) // % uint32(cms.M) to fit into the table

		cms.T[i][j]++
	}

}

func (cms *CountMinSketch) SearchMin(element string) uint {

	var result uint = math.MaxUint

	for i := 0; i < int(cms.k); i++ {
		cms.hashFunctions[i].Reset()
		cms.hashFunctions[i].Write([]byte(element))
		j := cms.hashFunctions[i].Sum32() % uint32(cms.m)

		if result > cms.T[i][j] {
			result = cms.T[i][j]
		}

	}

	return result

}

func CalculateM(epsilon float64) uint {
	return uint(math.Ceil(math.E / epsilon))
}

func CalculateK(delta float64) uint {
	return uint(math.Ceil(math.Log(math.E / delta)))
}

func (cms CountMinSketch) SerializeCountMinSketch(name string) string {
	name = "data/ds/cms/usertable-" + name + "-CMS.db"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		panic(err)
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(cms)

	if err != nil {
		panic(err)
	}

	err = file.Close()

	if err != nil {
		panic(err)
	}

	return name
}

func DeserializeCountMinSketch(name string) *CountMinSketch {
	file, err := os.OpenFile(name, os.O_RDWR, 0777)

	if err != nil {
		panic(err)
	}

	cms := CountMinSketch{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cms)

	if err != nil {
		panic(err)
	}

	cms.hashFunctions, cms.ts = utils.СreateHashFunctionsWithTS(cms.k, cms.ts)
	file.Close()
	return &cms
}

//func main() {
//
//	test := CreateCountMinSketch(0.01, 0.01)
//	testString := "BlumFilter2.0"
//	test.Addiction(testString)
//
//	minimum := test.SearchMin(testString)
//
//	fmt.Println(minimum)
//}

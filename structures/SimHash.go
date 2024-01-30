package structures

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

type SimHash struct {
	m           int   // Words-amount
	fingerprint []int // Control array
	text        string
	words       map[string]int
}

func (shash *SimHash) FingerprintInitialization() {

	words_arr := strings.Split(shash.text, " ")

	words := make(map[string]int)

	for _, word := range words_arr {
		words[word] += 1
	}

	shash.words = words
	shash.m = len(shash.words)

	table := make([][]string, shash.m)

	for m := range table { // Making 256 b space for hash
		table[m] = make([]string, 256)
	}

	i := 0
	for word, _ := range shash.words {

		s := GetBinaryString(GetSHA256Hash(word))

		for j := 0; j < len(s); j++ {
			if string(s[j]) == "0" {
				table[i][j] = "-1"
			} else {
				table[i][j] = "1"
			}
		}
		i += 1
	}

	fingerprint := make([]int, 256)

	i = 0
	for _, amount := range shash.words {

		for k := 0; k < len(table[i]); k++ {
			n, _ := strconv.Atoi(table[i][k])
			fingerprint[k] += n * amount
		}

		i += 1
	}

	for i, cell := range fingerprint {

		if cell <= 0 {
			fingerprint[i] = 0
		} else {
			fingerprint[i] = 1
		}

	}

	shash.fingerprint = fingerprint

}

func GetBinaryString(s string) string {
	var res []byte
	for _, c := range s {
		res = strconv.AppendInt(res, int64(c), 2)
	}
	return string(res)
}

func GetSHA256Hash(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Distance(h1, h2 SimHash) int {
	distance := 0

	for i, e := range h1.fingerprint {
		if e != h2.fingerprint[i] {
			distance += 1
		}
	}

	return distance
}

//func main() {
//
//	h1 := SimHash{text: "Hello, my name is Alberto"}
//	h1.FingerprintInitialization()
//
//	h2 := SimHash{text: "Hello, my name is Antonio"}
//	h2.FingerprintInitialization()
//
//	d := Distance(h1, h2)
//
//	fmt.Println("Distance ", d)
//}

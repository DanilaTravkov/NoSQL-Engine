package structures

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Memtable struct {
	size        uint
	threshold   uint
	currentSize uint
	lsm         [2]int
	skipList    *SkipList
	wal         *WAL
}

func CreateMemtable(data map[string]int, fromYaml bool) *Memtable {
	var (
		walSegmentSize, walLWM, memtableSize, memtableThreshold, lsmMaxLevel, lsmMergeThreshold, skiplistMaxHeight int = 5, 3, 10, 70, 3, 2, 10
	)
	if fromYaml {
		walSegmentSize = data["wal_size"]
		walLWM = data["wal_lwm"]
		memtableSize = data["memtable_size"]
		memtableThreshold = data["memtable_threshold"]
		skiplistMaxHeight = data["skiplist_max_height"]
	}
	memtable := Memtable{}
	memtable.size = uint(memtableSize)
	memtable.threshold = uint(memtableThreshold)
	memtable.lsm = [2]int{lsmMaxLevel, lsmMergeThreshold}
	insideWal := CreateWAL(uint8(walSegmentSize), uint8(walLWM))
	memtable.wal = insideWal
	memtable.skipList = CreateSkipList(skiplistMaxHeight, 0, 0)
	memtable.RecreateWALandSkipList()
	return &memtable
}

func (m *Memtable) RecreateWALandSkipList() {
	writtenSegments := 0

	files, _ := ioutil.ReadDir("data/ds/wal/segments")
	for _, f := range files {
		str := f.Name()
		file, _ := os.OpenFile("data/ds/wal/segments/"+str, os.O_RDONLY, 0777)
		file.Seek(0, 0)
		writtenSegments = 0
		for {
			crc := make([]byte, 4)
			_, err := file.Read(crc)
			c := binary.LittleEndian.Uint32(crc)

			if err == io.EOF {
				break
			}

			writtenSegments += 1

			wt := make([]byte, 8)
			file.Read(wt)
			walTimestamp := binary.LittleEndian.Uint64(wt)

			whatToDo := make([]byte, 1)
			file.Read(whatToDo)

			keySize := make([]byte, 8)
			file.Read(keySize)
			n := binary.LittleEndian.Uint64(keySize)

			valueSize := make([]byte, 8)
			file.Read(valueSize)
			mm := binary.LittleEndian.Uint64(valueSize)

			key := make([]byte, n)
			file.Read(key)
			value := make([]byte, mm)
			file.Read(value)

			if CRC32(value) != c {
				panic("Error")
			}
			errNew, isNew := m.skipList.UpdateTimestamp(string(key), value, int64(walTimestamp), whatToDo[0])
			if isNew {
				m.currentSize += 1
			}
			if errNew != nil {
				m.skipList.AddDeletedElement(string(key), value, int64(walTimestamp))
			}
		}
		m.wal.SetCurrentSize(uint8(writtenSegments))
		file.Close()
	}
}

func (m *Memtable) Write(key string, value []byte) bool {
	success := m.wal.AddElement(key, value)
	if success {
		err, isNew := m.skipList.AddElement(key, value)
		if err != nil {
			return false
		}
		if isNew {
			m.currentSize += 1
		}
		if m.currentSize*100 >= m.size*m.threshold {
			m.currentSize = 0
			m.Flush()
		}
		return true
	}
	return false
}

func (m *Memtable) Delete(key string, value []byte, isSomewhere bool) bool {
	if isSomewhere {
		success := m.wal.DeleteElement(key, value)
		if success {
			s := m.skipList.RemoveElement(key)

			if s == 0 {
				return true
			} else if s == 1 {
				now := time.Now()
				timestamp := now.Unix()
				err := m.skipList.AddDeletedElement(key, value, timestamp)
				if err != nil {
					panic(err)
				}
				m.currentSize += 1
				return true
			}
		}
	}
	return false
}

func (m *Memtable) Get(key string) (bool, bool, []byte) {
	valueNode := m.skipList.FindElement(key)
	if valueNode != nil {
		if valueNode.GetTombstone() == false {
			return true, false, valueNode.GetValue()
		} else {
			return true, true, valueNode.GetValue()
		}

	}
	return false, false, []byte("Nothing")
}

func (m *Memtable) Compactions(whatLvl int) {
	current := 0
	for current < m.lsm[1] {

		current += 1
		file1, _ := os.OpenFile("data/ds/data/usertable-lvl="+strconv.Itoa(whatLvl)+"-gen="+strconv.Itoa(current)+"-Data.db", os.O_RDONLY, 0777)
		index1, _ := os.OpenFile("data/ds/index/usertable-lvl="+strconv.Itoa(whatLvl)+"-gen="+strconv.Itoa(current)+"-Index.db", os.O_RDONLY, 0777)
		current += 1
		file2, _ := os.OpenFile("data/ds/data/usertable-lvl="+strconv.Itoa(whatLvl)+"-gen="+strconv.Itoa(current)+"-Data.db", os.O_RDONLY, 0777)
		index2, _ := os.OpenFile("data/ds/index/usertable-lvl="+strconv.Itoa(whatLvl)+"-gen="+strconv.Itoa(current)+"-Index.db", os.O_RDONLY, 0777)

		n := FindLSMGeneration(whatLvl + 1)

		index1Size, _ := index1.Stat()
		index2Size, _ := index2.Stat()

		newMerkle := MerkleTree{}
		newBloom := CreateBloomFilter(int(index1Size.Size())+int(index2Size.Size()), 0.01)
		newSkipList := CreateSkipList(10, 0, 0)
		newCms := CreateCountMinSketch(0.01, 0.01)
		newHll := CreateHLL(uint8(4))

		for {
			iLenBytes := make([]byte, 8)
			index1.Read(iLenBytes)
			iLen := binary.LittleEndian.Uint64(iLenBytes)

			jLenBytes := make([]byte, 8)
			index2.Read(jLenBytes)
			jLen := binary.LittleEndian.Uint64(jLenBytes)

			i := make([]byte, iLen+8)
			j := make([]byte, jLen+8)
			_, err1 := index1.Read(i)
			_, err2 := index2.Read(j)

			if err1 == io.EOF {
				index2.Seek(-16-int64(jLen), 1)
				break
			} else if err2 == io.EOF {
				index1.Seek(-16-int64(iLen), 1)
				break
			}

			offset1 := binary.LittleEndian.Uint64(i[iLen:])
			offset2 := binary.LittleEndian.Uint64(j[jLen:])

			file1.Seek(int64(offset1), 0)
			file2.Seek(int64(offset2), 0)

			writeTime1, tombstone1, key1, value1 := PrepareData(file1)
			writeTime2, tombstone2, key2, value2 := PrepareData(file2)

			if string(key1) == string(key2) {
				if tombstone2[0] == 0 {
					AddOnNextLevel(&newMerkle, &newBloom, newSkipList, newCms, newHll, writeTime2, tombstone2, key2, value2)
				}
			} else {
				if string(key1) < string(key2) {
					AddOnNextLevel(&newMerkle, &newBloom, newSkipList, newCms, newHll, writeTime1, tombstone1, key1, value1)
					index2.Seek(-16-int64(jLen), 1)
				} else {
					AddOnNextLevel(&newMerkle, &newBloom, newSkipList, newCms, newHll, writeTime2, tombstone2, key2, value2)
					index1.Seek(-16-int64(iLen), 1)
				}
			}
		}
		for {
			iLenBytes := make([]byte, 8)
			index1.Read(iLenBytes)
			iLen := binary.LittleEndian.Uint64(iLenBytes)

			i := make([]byte, iLen+8)
			_, err1 := index1.Read(i)
			if err1 == io.EOF {
				break
			}

			offset1 := binary.LittleEndian.Uint64(i[iLen:])
			file1.Seek(int64(offset1), 0)
			writeTime1, tombstone1, key1, value1 := PrepareData(file1)
			AddOnNextLevel(&newMerkle, &newBloom, newSkipList, newCms, newHll, writeTime1, tombstone1, key1, value1)
		}
		for {
			jLenBytes := make([]byte, 8)
			index2.Read(jLenBytes)

			jLen := binary.LittleEndian.Uint64(jLenBytes)

			j := make([]byte, jLen+8)
			_, err1 := index2.Read(j)
			if err1 == io.EOF {
				break
			}

			offset2 := binary.LittleEndian.Uint64(j[jLen:])
			file2.Seek(int64(offset2), 0)
			writeTime2, tombstone2, key2, value2 := PrepareData(file2)
			AddOnNextLevel(&newMerkle, &newBloom, newSkipList, newCms, newHll, writeTime2, tombstone2, key2, value2)
		}

		index1.Close()
		index2.Close()
		file1.Close()
		file2.Close()

		removeOldFiles(whatLvl, current)

		// usertable-lvl=LVL-gen=GEN-Filter.db
		newBloom.SerializeBloomFilter(n+1, whatLvl+1)
		// usertable-lvl=LVL-gen=GEN-Metadata.db
		newMerkle.CreateTree()
		newMerkle.Serialize(n+1, whatLvl+1)

		elements := newSkipList.LastLevel()
		// usertable-lvl=LVL-gen=GEN-Data.db
		// usertable-lvl=LVL-gen=GEN-Index.db
		// usertable-lvl=LVL-gen=GEN-Summary.db
		createSSTable(elements, n+1, whatLvl+1)

		// usertable-lvl=LVL-gen=GEN-TOC.txt
		file, err := os.Create("data/ds/toc/usertable-lvl=" + strconv.Itoa(whatLvl+1) + "-gen=" + strconv.Itoa(n+1) + "-TOC.txt")
		if err != nil {
			panic(err)
		}
		_, err = file.Write([]byte("bloomFilter/usertable-lvl=" + strconv.Itoa(whatLvl+1) + "-gen=" + strconv.Itoa(n+1) + "-Filter.db\nmerkleTree/usertable-lvl=" + strconv.Itoa(whatLvl+1) + "-gen=" + strconv.Itoa(n+1) + "-Metadata.db\ndata/ds/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(n+1) + "-Data.db\nindex/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(n+1) + "-Index.db\nsummary/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(n+1) + "-Summary.db\ntoc/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(n+1) + "-TOC.txt\n"))
		if err != nil {
			panic(err)
		}
		file.Close()
	}
	nn := FindLSMGeneration(whatLvl + 1)
	if nn == m.lsm[1] && whatLvl+1 < m.lsm[0] {
		m.Compactions(whatLvl + 1)
	}
}

func AddOnNextLevel(newMerkle *MerkleTree, newBloom *BloomFilter, newSkipList *SkipList,
	newCms *CountMinSketch, newHll *HLL, writeTime uint64, tombstone []byte, key []byte, value []byte) {
	newCms.Addiction(string(key))
	newHll.Add(string(key))
	newMerkle.AddElement(value)
	newBloom.AddElement(string(key))
	if tombstone[0] == 0 {
		newSkipList.AddElement(string(key), value)
	} else {
		newSkipList.AddDeletedElement(string(key), value, int64(writeTime))
	}

}
func PrepareData(file *os.File) (uint64, []byte, []byte, []byte) {
	crc1 := make([]byte, 4)
	file.Read(crc1)
	c1 := binary.LittleEndian.Uint32(crc1)
	timestamp := make([]byte, 8)
	file.Read(timestamp)
	writeTime := binary.LittleEndian.Uint64(timestamp)
	tombstone := make([]byte, 1)
	file.Read(tombstone)
	keySize := make([]byte, 8)
	file.Read(keySize)
	n := binary.LittleEndian.Uint64(keySize)
	valueSize := make([]byte, 8)
	file.Read(valueSize)
	m := binary.LittleEndian.Uint64(valueSize)
	key := make([]byte, n)
	file.Read(key)
	value := make([]byte, m)
	file.Read(value)
	if CRC32(value) != c1 {
		panic("NEEEEEEE")
	}
	return writeTime, tombstone, key, value
}

func (m *Memtable) Flush() {
	gen := findCurrentGeneration()
	elements := m.skipList.LastLevel()

	merkle := MerkleTree{}
	bloom := CreateBloomFilter(len(elements), 0.01)
	for _, el := range elements {
		merkle.AddElement(el.GetValue())
		bloom.AddElement(el.GetKey())
	}

	bloom.SerializeBloomFilter(gen+1, 1)

	merkle.CreateTree()
	merkle.Serialize(gen+1, 1)

	createSSTable(elements, gen+1, 1)

	file, err := os.Create("data/ds/toc/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-TOC.txt")
	if err != nil {
		panic(err)
	}
	_, err = file.Write([]byte("bloomFilter/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-Filter.db\nmerkleTree/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-Metadata.db\ndata/ds/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-Data.db\nindex/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-Index.db\nsummary/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-Summary.db\ntoc/usertable-lvl=" + strconv.Itoa(1) + "-gen=" + strconv.Itoa(gen+1) + "-TOC.txt\n"))
	if err != nil {
		panic(err)
	}
	file.Close()

	m.skipList = CreateSkipList(10, 0, 0)
	m.wal = m.wal.ResetWAL()

	if gen+1 == m.lsm[1] {
		m.Compactions(1)
	}
}

func createSSTable(elements []*SkipListNode, gen, lvl int) {
	var offset uint64 = 0
	var indexOffset uint64 = 0
	fileData, err1 := os.OpenFile("data/ds/data/usertable-lvl="+strconv.Itoa(lvl)+"-gen="+strconv.Itoa(gen)+"-Data.db", os.O_WRONLY|os.O_CREATE, 0777)
	if err1 != nil {
		panic(err1)
	}
	fileIndex, err2 := os.OpenFile("data/ds/index/usertable-lvl="+strconv.Itoa(lvl)+"-gen="+strconv.Itoa(gen)+"-Index.db", os.O_WRONLY|os.O_CREATE, 0777)
	if err2 != nil {
		panic(err2)
	}
	fileSummary, err3 := os.OpenFile("data/ds/summary/usertable-lvl="+strconv.Itoa(lvl)+"-gen="+strconv.Itoa(gen)+"-Summary.db", os.O_WRONLY|os.O_CREATE, 0777)
	if err3 != nil {
		panic(err3)
	}

	first := elements[0].GetKey()
	last := elements[len(elements)-1].GetKey()

	var firstSize uint64 = uint64(len(first))
	firstSize_final := make([]byte, 8)
	binary.LittleEndian.PutUint64(firstSize_final, firstSize)
	_, err := fileSummary.Write(firstSize_final)
	if err != nil {
		panic(err)
	}

	_, err = fileSummary.Write([]byte(first))
	if err != nil {
		panic(err)
	}

	var lastSize uint64 = uint64(len(last))
	lastSize_final := make([]byte, 8)
	binary.LittleEndian.PutUint64(lastSize_final, lastSize)
	_, err = fileSummary.Write(lastSize_final)
	if err != nil {
		panic(err)
	}

	_, err = fileSummary.Write([]byte(last))
	if err != nil {
		panic(err)
	}

	for _, element := range elements {
		// START - write to data
		crc := CRC32(element.GetValue())
		crc_final := make([]byte, 4)
		binary.LittleEndian.PutUint32(crc_final, crc)

		timestamp_final := make([]byte, 8)
		binary.LittleEndian.PutUint64(timestamp_final, uint64(element.GetTimestamp()))

		tombstone_final := make([]byte, 1)
		if element.GetTombstone() {
			tombstone_final[0] = 1
		}

		var keySize uint64 = uint64(len(element.GetKey()))
		keySize_final := make([]byte, 8)
		binary.LittleEndian.PutUint64(keySize_final, keySize)

		var valueSize uint64 = uint64(len(element.GetValue()))
		valueSize_final := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueSize_final, valueSize)

		key_final := element.GetKey()
		// write to data
		fileData.Write(crc_final)
		fileData.Write(timestamp_final)
		fileData.Write(tombstone_final)
		fileData.Write(keySize_final)
		fileData.Write(valueSize_final)
		fileData.Write([]byte(key_final))
		fileData.Write(element.GetValue())
		recordSize := 4 + 8 + 1 + 8 + 8 + keySize + valueSize

		// write to index
		offset_final := make([]byte, 8)
		binary.LittleEndian.PutUint64(offset_final, offset)
		fileIndex.Write(keySize_final)
		fileIndex.Write([]byte(key_final))
		fileIndex.Write(offset_final)
		offset += recordSize
		indexSize := keySize + 16

		// write summary elements (borders already written)
		fileSummary.Write(keySize_final)
		fileSummary.Write([]byte(key_final))
		index_offset_final := make([]byte, 8)
		binary.LittleEndian.PutUint64(index_offset_final, indexOffset)
		fileSummary.Write(index_offset_final)
		indexOffset += indexSize
	}

	fileData.Close()
	fileIndex.Close()
	fileSummary.Close()
}

func findCurrentGeneration() int {
	files, _ := ioutil.ReadDir("data/ds/data")
	maxName := 0
	for _, f := range files {
		str := f.Name()
		ok, _ := regexp.Match("usertable-lvl=1-gen=\\d+-Data.db", []byte(str))
		if ok {
			x := strings.Split(str, "-Data.db")
			x = strings.Split(x[0], "usertable-lvl=1-gen=")
			num, _ := strconv.Atoi(x[1])
			if num > maxName {
				maxName = num
			}
		}
	}
	return maxName
}

func FindLSMGeneration(whatLvl int) int {
	files, _ := ioutil.ReadDir("data/ds/data")
	maxName := 0
	for _, f := range files {
		str := f.Name()
		re := regexp.MustCompile("-gen=\\d+-Data.db")
		lvl := re.Split(str, -1)
		lvl = strings.Split(lvl[0], "usertable-lvl=")
		l, _ := strconv.Atoi(lvl[1])
		if l == whatLvl {
			re = regexp.MustCompile("usertable-lvl=\\d+-gen=")
			gen := re.Split(str, -1)
			gen = strings.Split(gen[1], "-Data.db")
			g, _ := strconv.Atoi(gen[0])
			if g > maxName {
				maxName = g
			}
		}
	}
	return maxName
}

func removeOldFiles(whatLvl, current int) {
	e := os.Remove("data/ds/index/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-Index.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/data/ds/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-Data.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/index/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-Index.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/data/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-Data.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/bloomFilter/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-Filter.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/bloomFilter/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-Filter.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/merkleTree/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-Metadata.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/merkleTree/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-Metadata.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/summary/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-Summary.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/summary/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-Summary.db")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/toc/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current-1) + "-TOC.txt")
	if e != nil {
		log.Fatal(e)
	}
	e = os.Remove("data/ds/toc/usertable-lvl=" + strconv.Itoa(whatLvl) + "-gen=" + strconv.Itoa(current) + "-TOC.txt")
	if e != nil {
		log.Fatal(e)
	}
}

func (m *Memtable) Finish() {
	m.wal.Finish()
}

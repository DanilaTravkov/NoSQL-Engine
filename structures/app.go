package structures

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	tokenBucket *TokenBucket
}

type App struct {
	memtable    *Memtable
	cache       *Cache
	tokenBucket *TokenBucket
	data        map[string]int
	user        User
	cms         map[string]string
	hll         map[string]string
}

func CreateApp() App {
	app := App{}
	app.data = make(map[string]int)
	fromYaml := false
	yfile, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		yaml.Unmarshal(yfile, &app.data)
		fromYaml = true
	} else {
		app.data["wal_size"] = 5
		app.data["wal_lwm"] = 3
		app.data["memtable_size"] = 10
		app.data["memtable_threshold"] = 70
		app.data["cache_max_size"] = 5
		app.data["lsm_max_lvl"] = 3
		app.data["lsm_merge_threshold"] = 2
		app.data["skiplist_max_height"] = 10
		app.data["hll_precision"] = 4
		app.data["tokenbucket_size"] = 200
		app.data["tokenbucket_interval"] = 60
		app.data["cmsEpsilon"] = 1
		app.data["cmsDelta"] = 1
	}
	app.memtable = CreateMemtable(app.data, fromYaml)
	app.cache = CreateCache(uint32(app.data["cache_max_size"]))
	app.hll = make(map[string]string)
	app.cms = make(map[string]string)
	app.createNewHLL("default", uint8(app.data["hll_precision"]))
	app.createNewCMS("default", float64(app.data["cmsEpsilon"])*0.01, float64(app.data["cmsDelta"])*0.01)

	return app
}

func (app *App) RunApp(amateur bool) {
	if !amateur {
		http.HandleFunc("/login/", app.login)
		http.HandleFunc("/data/ds/", app.users)
		http.HandleFunc("/", app.index)

		port := os.Getenv("PORT")
		if port == "" {
			port = "9000"
		}
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			panic(err)
		}
	} else {
		for true {
			fmt.Print("Choose on option:\n 1) Login \n 2) Register \n 3) Exit \n>> ")
			var option string
			fmt.Scanln(&option)
			if strings.Replace(option, ")", "", 1) == "1" {
				fmt.Print("Enter the username\n >> ")
				var username string
				fmt.Scanln(&username)

				fmt.Print("Enter the password\n >> ")
				var password string
				fmt.Scanln(&password)

				if app._login(username, password) {
					app.options()
				}
			} else if strings.Replace(option, ")", "", 1) == "2" {
				fmt.Print("Enter the username\n >> ")
				var username string
				fmt.Scanln(&username)

				fmt.Print("Enter the password\n >> ")
				var password string
				fmt.Scanln(&password)

				if !app._login(username, password) {
					file, err := os.OpenFile("data/ds/users/users.csv", os.O_APPEND|os.O_CREATE, 0777)
					if err != nil {
						panic(err)
					}

					_, err = file.WriteString(username + "," + password + "," + strconv.Itoa(app.data["tokenbucket_size"]) + "," + strconv.Itoa(int(time.Now().Unix())) + "\n")
					if err != nil {
						panic(err)
					}
					file.Close()
				} else {
					fmt.Println("\nError: this valuse is not unique\n")
				}
			} else if strings.Replace(option, ")", "", 1) == "3" {
				break
			} else {
				fmt.Println("\nError: you don't have this option. Try again\n")
			}
		}
	}
}

func (app *App) options() {
	for true {
		fmt.Print("Choose one option:\n 1) Put \n 2) Get \n 3) Delete \n 4) PutSpecial \n 5) Test \n 6) Logout \n >> ")
		var option string
		fmt.Scanln(&option)
		if strings.Replace(option, ")", "", 1) == "1" {
			fmt.Print("Enter the key\n >> ")
			var key string
			fmt.Scanln(&key)
			fmt.Print("Enter the value\n >> ")
			var value string
			fmt.Scanln(&value)
			ok := app._put(key, []byte(value))
			if ok {
				fmt.Println("\nSuccesfully!\n")
			} else {
				fmt.Println("\nError: wrong addiction\n")
			}
		} else if strings.Replace(option, ")", "", 1) == "2" {
			fmt.Print("Enter the key\n >> ")
			var key string
			fmt.Scanln(&key)
			ok, value := app._get(key)
			if ok {
				fmt.Println("\nAmount of the key " + key + " is " + string(value) + "\n")
			} else {
				fmt.Println(string(value))
			}
		} else if strings.Replace(option, ")", "", 1) == "3" {
			fmt.Print("Enter the key\n >> ")
			var key string
			fmt.Scanln(&key)
			fmt.Print("Enter the value\n >> ")
			var value string
			fmt.Scanln(&value)
			ok := app._delete(key, []byte(value))
			if ok {
				fmt.Println("\nSuccesfully!\n")
			} else {
				fmt.Println("\nError: wrong deletion\n")
			}
		} else if strings.Replace(option, ")", "", 1) == "4" {
			fmt.Print("Enter the key\n >> ")
			var key string
			fmt.Scanln(&key)
			fmt.Print("Enter the value\n >> ")
			var value string
			fmt.Scanln(&value)
			fmt.Print("Enter the category\n >> ")
			var whichOne string
			fmt.Scanln(&whichOne)
			ok := app._putSpecial(key, []byte(value), whichOne)
			if ok {
				fmt.Println("\nSuccesfully!\n")
			} else {
				fmt.Println("\nError: wrong addiction\n")
			}
		} else if strings.Replace(option, ")", "", 1) == "5" {
			ok := app.Test()
			if !ok {
				fmt.Println("\nError: Tests\n")
			}
		} else if strings.Replace(option, ")", "", 1) == "6" {
			break
		} else {
			fmt.Println("\nError: choose another option\n")
		}
		app.updateUsersFile()
	}
}

func (app *App) StopApp() {
	app.memtable.Finish()
	os.Exit(0)
}

func (app *App) createNewHLL(key string, precision uint8) bool {
	_, ok := app.hll[key]
	if ok {
		fmt.Println("HLL with this data is already exists!")
		return false
	}

	newHLL := CreateHLL(precision)
	name := newHLL.SerializeHyperLogLog(key)
	app.hll[key] = name

	return true
}

func (app *App) createNewCMS(key string, epsilon, delta float64) bool {
	_, ok := app.cms[key]
	if ok {
		fmt.Println("CMS with this data is already exists!")
		return false
	}

	newCMS := CreateCountMinSketch(epsilon, delta)
	name := newCMS.SerializeCountMinSketch(key)
	app.cms[key] = name
	return true
}

func (app *App) addToSpecialHLL(key string, whichOne string) bool {
	name, ok := app.hll[whichOne]
	if !ok {
		app.createNewHLL(whichOne, uint8(app.data["hll_precision"]))
	}
	name, _ = app.hll[whichOne]
	sHLL := DeserializeHyperLogLog(name)
	sHLL.Add(key)
	sHLL.SerializeHyperLogLog(whichOne)
	return true
}

func (app *App) addToSpecialCMS(key string, whichOne string) bool {
	name, ok := app.cms[whichOne]
	if !ok {
		app.createNewCMS(whichOne, float64(app.data["cmsEpsilon"])*0.01, float64(app.data["cmsDelta"])*0.01)
	}
	name, _ = app.cms[whichOne]
	sCMS := DeserializeCountMinSketch(name)
	sCMS.Addiction(key)
	sCMS.SerializeCountMinSketch(whichOne)
	return true
}

func (app *App) getEstimateFromSpecialHLL(whichOne string) float64 {
	sHLL := DeserializeHyperLogLog(app.hll[whichOne])
	return sHLL.Estimate()
}

func (app *App) getFrequencyFromSpecialCMS(key string, whichOne string) uint {
	sCMS := DeserializeCountMinSketch(app.cms[whichOne])
	return sCMS.SearchMin(key)
}

func (app *App) _putSpecial(key string, value []byte, whichOne string) bool {
	if app.tokenBucket.Update() {
		ok := app._put(key, value)
		if ok {
			app.addToSpecialHLL(key, whichOne)
			app.addToSpecialCMS(key, whichOne)
			return true
		}
		return false
	}
	fmt.Println("Error: You have reached the maximum number of requests. Please try again later.")
	return false
}

func (app *App) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	jsonBytes, err := json.Marshal("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (app *App) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "GET" {
		tokens := strings.Split(r.URL.String(), "/")
		if len(tokens) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		info := strings.Split(tokens[2], ",")
		if len(info) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ok := app._login(info[0], info[1])

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		app.updateUsersFile()
		return
	} else {
		tokens := strings.Split(r.URL.String(), "/")
		if len(tokens) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		info := strings.Split(tokens[2], ",")
		if len(info) != 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !app._login(info[0], info[1]) {
			file, err := os.OpenFile("data/ds/users/users.csv", os.O_APPEND|os.O_WRONLY, 0777)
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(info[0] + "," + info[1] + "," + strconv.Itoa(app.data["tokenbucket_size"]) + "," + strconv.Itoa(int(time.Now().Unix())) + "\n")
			if err != nil {
				panic(err)
			}
			file.Close()
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func (app *App) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	tokens := strings.Split(r.URL.String(), "/")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ok, value := app._get(tokens[2])

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(value)
}

func (app *App) put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	tokens := strings.Split(r.URL.String(), "/")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	info := strings.Split(tokens[2], ",")
	if len(info) != 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ok := app._put(info[0], []byte(info[1]))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *App) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	tokens := strings.Split(r.URL.String(), "/")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	info := strings.Split(tokens[2], ",")
	if len(info) != 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ok := app._delete(info[0], []byte(info[1]))
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *App) users(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch r.Method {
	case "GET":
		app.get(w, r)
		app.updateUsersFile()
		return
	case "POST", "PUT":
		app.put(w, r)
		app.updateUsersFile()
		return
	case "OPTIONS":
		app.delete(w, r)
		app.updateUsersFile()
		return
	default:
		return
	}
}

func (app *App) _put(key string, value []byte) bool {
	if app.tokenBucket.Update() {
		ok := app.memtable.Write(key, value)
		if ok {
			cms := DeserializeCountMinSketch(app.cms["default"])
			cms.Addiction(key)
			cms.SerializeCountMinSketch("default")
			hll := DeserializeHyperLogLog(app.hll["default"])
			hll.Add(key)
			hll.SerializeHyperLogLog("default")
		}
		return ok
	}
	fmt.Println("Erorr: You have reached the maximum number of requests. Please try again later.")
	return false
}

func (app *App) _get(key string) (bool, []byte) {
	if app.tokenBucket.Update() {
		var value []byte
		var isThere, deleted bool
		isThere, deleted, value = app.memtable.Get(key)
		if isThere {
			if deleted {
				return false, []byte("The data has been logically deleted.")
			} else {
				app.cache.AddElement(key, value)
				return true, value
			}
		}

		isThere, value = app.cache.GetElement(key)
		if isThere {
			app.cache.AddElement(key, value)
			return true, value
		}

		for i := 1; i <= app.data["lsm_max_lvl"]; i++ {
			maxGen := FindLSMGeneration(i)
			for j := 1; j <= maxGen; j++ {

				gen := j
				if i == app.data["lsm_max_lvl"] {
					j = maxGen - j + 1
				}
				bloomFilter := DeserializeBloomFilter(j, i)
				isThere = bloomFilter.IsElementInBloomFilter(key)
				if isThere {

					fileSummary, _ := os.OpenFile("data/ds/summary/usertable-lvl="+strconv.Itoa(i)+"-gen="+strconv.Itoa(j)+"-Summary.db", os.O_RDONLY, 0777)

					firstSizeBytes := make([]byte, 8)
					fileSummary.Read(firstSizeBytes)
					firstSize := binary.LittleEndian.Uint64(firstSizeBytes)

					firstIndexBytes := make([]byte, firstSize)
					fileSummary.Read(firstIndexBytes)

					lastSizeBytes := make([]byte, 8)
					fileSummary.Read(lastSizeBytes)
					lastSize := binary.LittleEndian.Uint64(lastSizeBytes)

					lastIndexBytes := make([]byte, lastSize)
					fileSummary.Read(lastIndexBytes)

					if key >= string(firstIndexBytes) && key <= string(lastIndexBytes) {
						summeryStructure := make(map[string]uint64)
						for {
							keyLenBytes := make([]byte, 8)
							_, err := fileSummary.Read(keyLenBytes)
							if err == io.EOF {
								break
							}
							keyLen := binary.LittleEndian.Uint64(keyLenBytes)

							buff := make([]byte, keyLen+8)
							fileSummary.Read(buff)
							keyBytes := buff[:keyLen]
							indexPosition := binary.LittleEndian.Uint64(buff[keyLen:])
							summeryStructure[string(keyBytes)] = indexPosition
						}

						indexPosition, existInMap := summeryStructure[key]
						if existInMap {

							fileIndex, _ := os.OpenFile("data/ds/index/usertable-lvl="+strconv.Itoa(i)+"-gen="+strconv.Itoa(j)+"-Index.db", os.O_RDONLY, 0777)
							fileIndex.Seek(int64(indexPosition), 0)

							keyLenIndexBytes := make([]byte, 8)
							fileIndex.Read(keyLenIndexBytes)
							keyLenIndex := binary.LittleEndian.Uint64(keyLenIndexBytes)

							buff2 := make([]byte, keyLenIndex+8)
							fileIndex.Read(buff2)
							dataPosition := binary.LittleEndian.Uint64(buff2[keyLenIndex:])

							fileIndex.Close()

							fileData, _ := os.OpenFile("data/ds/data/ds/usertable-lvl="+strconv.Itoa(i)+"-gen="+strconv.Itoa(j)+"-Data.db", os.O_RDONLY, 0777)
							fileData.Seek(int64(dataPosition), 0)

							crc := make([]byte, 4)
							fileData.Read(crc)
							c := binary.LittleEndian.Uint32(crc)

							fileData.Seek(8, 1)

							whatToDo := make([]byte, 1)
							fileData.Read(whatToDo)
							if whatToDo[0] == 1 {

								fileSummary.Close()
								fileData.Close()

								return false, []byte("The data has been logically deleted.")
							}

							keySize := make([]byte, 8)
							fileData.Read(keySize)
							n := binary.LittleEndian.Uint64(keySize)

							valueSize := make([]byte, 8)
							fileData.Read(valueSize)
							mm := binary.LittleEndian.Uint64(valueSize)

							keyData := make([]byte, n)
							fileData.Read(keyData)
							value = make([]byte, mm)
							fileData.Read(value)
							if CRC32(value) != c {
								panic("It won't budge.")
							}
							fileSummary.Close()
							fileData.Close()
							app.cache.AddElement(key, value)
							fileData.Close()
							fileSummary.Close()
							return true, value
						}
					}
					fileSummary.Close()
				}
				j = gen
			}
		}
		return false, []byte("It doesn't exist.")
	}
	return false, []byte("Error: You have reached the maximum number of requests. Please try again later.")
}

func (app *App) _delete(key string, value []byte) bool {
	if app.tokenBucket.Update() {
		answer, _ := app._get(key)
		app.cache.RemoveElement(key)
		return app.memtable.Delete(key, value, answer)
	}
	fmt.Println("Error: You have reached the maximum number of requests. Please try again later.")
	return false
}

func (app *App) _login(username, password string) bool {
	file, err := os.OpenFile("data/ds/users/users.csv", os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		items := strings.Split(scanner.Text(), ",")
		if username == items[0] && password == items[1] {
			app.user.Username = username
			app.user.Password = password
			tokensLeft, _ := strconv.Atoi(items[2])
			lastReset, _ := strconv.Atoi(items[3])
			app.user.tokenBucket = CreateTokenBucket(app.data["tokenbucket_size"], app.data["tokenbucket_interval"], tokensLeft, int64(lastReset))
			app.tokenBucket = app.user.tokenBucket
			return true
		}
	}
	return false
}

func (app *App) updateUsersFile() {
	file, err := os.OpenFile("data/ds/users/users.csv", os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	list := make([]byte, 0)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), app.user.Username) {
			list = append(list, app.user.Username+","+app.user.Password+","+strconv.Itoa(app.user.tokenBucket.GetTokensLeft())+","+strconv.Itoa(int(app.user.tokenBucket.GetLastReset()))+"\n"...)
		} else {
			list = append(list, []byte(scanner.Text()+"\n")...)
		}
	}
	file.Close()

	file, err = os.OpenFile("data/ds/users/users.csv", os.O_RDWR, 0777)
	file.WriteString(string(list))
	file.Close()

}

func (app *App) Test() bool {
	app._put("key01", []byte("data1"))
	app._put("key02", []byte("data2"))
	app._put("key03", []byte("data3"))
	app._put("key04", []byte("data4"))
	app._put("key05", []byte("data5"))
	app._put("key06", []byte("data6"))
	app._put("key07", []byte("data7"))

	app._put("key08", []byte("data8"))
	app._delete("key03", []byte("data3"))
	app._put("key10", []byte("data10"))
	app._put("key11", []byte("data11"))
	app._put("key12", []byte("data12"))
	app._put("key13", []byte("data13"))
	app._put("key14", []byte("data14"))

	app._delete("key15", []byte("data15"))
	app._delete("key05", []byte("data5"))
	app._delete("key06", []byte("data6"))
	app._put("key18", []byte("data18"))
	app._delete("key08", []byte("data08"))
	app._put("key20", []byte("data20"))
	app._put("key21", []byte("data21"))
	app._put("key30", []byte("data30"))

	app._put("key23", []byte("data23"))
	app._put("key24", []byte("data24"))
	app._put("key25", []byte("data25"))
	app._put("key26", []byte("data26"))
	app._put("key27", []byte("data27"))
	app._put("key01", []byte("data01"))
	app._put("key28", []byte("data28"))

	app._put("key31", []byte("data31"))
	app._put("key32", []byte("data32"))
	app._put("key33", []byte("data33"))
	app._put("key34", []byte("data34"))
	app._put("key35", []byte("data35"))
	app._put("key36", []byte("data36"))
	app._put("key37", []byte("data37"))

	app._put("key08", []byte("data8"))
	app._delete("key07", []byte("data7"))
	app._put("key10", []byte("data10"))
	app._put("key11", []byte("data11"))
	app._put("key12", []byte("data12"))
	app._put("key13", []byte("data13"))
	app._put("key14", []byte("data14"))

	app._delete("key15", []byte("data15"))
	app._delete("key05", []byte("data5"))
	app._delete("key06", []byte("data6"))
	app._put("key18", []byte("data18"))
	app._delete("key08", []byte("data08"))
	app._put("key20", []byte("data20"))
	app._put("key21", []byte("data21"))
	app._put("key30", []byte("data30"))

	app._put("key23", []byte("data23"))
	app._put("key24", []byte("data24"))
	app._put("key25", []byte("data25"))
	app._put("key26", []byte("data26"))
	app._put("key27", []byte("data27"))
	app._put("key01", []byte("data01"))
	app._put("key28", []byte("data28"))

	app._put("key01", []byte("data1"))
	app._put("key02", []byte("new_data2"))
	app._put("key03", []byte("data3"))
	app._put("key04", []byte("data4"))
	app._put("key05", []byte("data5"))
	app._put("key06", []byte("data6"))
	app._put("key07", []byte("data7"))

	app._put("key08", []byte("data8"))
	app._delete("key03", []byte("data3"))
	app._put("key10", []byte("data10"))
	app._put("key11", []byte("data11"))
	app._put("key12", []byte("data12"))
	app._put("key13", []byte("data13"))
	app._put("key14", []byte("data14"))

	app._delete("key15", []byte("data15"))
	app._delete("key05", []byte("data5"))
	app._delete("key06", []byte("data6"))
	app._put("key18", []byte("data18"))
	app._delete("key08", []byte("data08"))
	app._put("key20", []byte("data20"))
	app._put("key21", []byte("data21"))
	app._put("key30", []byte("data30"))

	app._put("key23", []byte("data23"))
	app._put("key24", []byte("data24"))
	app._put("key25", []byte("data25"))
	app._put("key26", []byte("data26"))
	app._put("key27", []byte("data27"))
	app._put("key01", []byte("data01"))
	app._put("key28", []byte("data28"))

	_, value := app._get("key01")
	fmt.Println(string(value))
	_, value = app._get("key02")
	fmt.Println(string(value))
	_, value = app._get("key03")
	fmt.Println(string(value))
	_, value = app._get("key04")
	fmt.Println(string(value))
	_, value = app._get("key05")
	fmt.Println(string(value))
	_, value = app._get("key06")
	fmt.Println(string(value))
	_, value = app._get("key07")
	fmt.Println(string(value))
	_, value = app._get("key08")
	fmt.Println(string(value))
	_, value = app._get("key09")
	fmt.Println(string(value))
	_, value = app._get("key10")
	fmt.Println(string(value))
	_, value = app._get("key11")
	fmt.Println(string(value))
	_, value = app._get("key12")
	fmt.Println(string(value))
	_, value = app._get("key13")
	fmt.Println(string(value))
	_, value = app._get("key14")
	fmt.Println(string(value))
	_, value = app._get("key15")
	fmt.Println(string(value))
	_, value = app._get("key16")
	fmt.Println(string(value))
	_, value = app._get("key17")
	fmt.Println(string(value))
	_, value = app._get("key18")
	fmt.Println(string(value))
	_, value = app._get("key19")
	fmt.Println(string(value))
	_, value = app._get("key20")
	fmt.Println(string(value))
	_, value = app._get("key21")
	fmt.Println(string(value))
	_, value = app._get("key22")
	fmt.Println(string(value))
	_, value = app._get("key23")
	fmt.Println(string(value))
	_, value = app._get("key24")
	fmt.Println(string(value))
	_, value = app._get("key25")
	fmt.Println(string(value))
	_, value = app._get("key26")
	fmt.Println(string(value))
	_, value = app._get("key27")
	fmt.Println(string(value))
	_, value = app._get("key28")
	fmt.Println(string(value))
	_, value = app._get("key29")
	fmt.Println(string(value))
	_, value = app._get("key30")
	fmt.Println(string(value))
	_, value = app._get("key31")
	fmt.Println(string(value))
	_, value = app._get("key32")
	fmt.Println(string(value))
	_, value = app._get("key33")
	fmt.Println(string(value))
	_, value = app._get("key34")
	fmt.Println(string(value))
	_, value = app._get("key35")
	fmt.Println(string(value))
	_, value = app._get("key36")
	fmt.Println(string(value))
	_, value = app._get("key37")
	fmt.Println(string(value))

	app._putSpecial("tasa", []byte("tasa"), "new")

	return true
}

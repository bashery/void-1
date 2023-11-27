package kvlite

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Path string
}

type index struct {
	// location format is :
	// "i <key> <at> <size> <page-name>"
	// "i 0 199 45 0"
	key  string
	page string
	at   int64
	size int
}

type Database struct {
	name   string
	index  int64
	lastat int64
	pages  map[string]*os.File
	indexs map[string]index
	afile  string
	path   string
}

// Set inserts new or update exist value
func (db *Database) Set(key, value string) {

	size := len(value)

	// TODO use string builder to reduce memory consomption
	location := "\ni " + key + " " + fmt.Sprint(db.lastat) + " " + fmt.Sprint(size) + " 0\n"

	db.pages[db.afile].Write([]byte(value + location))

	// indexs
	db.indexs[key] = index{at: db.lastat, size: size, page: "0"}

	db.lastat += int64(size) + int64(len(location))
}

// Get data by key
func (db *Database) Get(key string) string {

	// location format is :
	// "i <key> <at> <size> <page-name>"
	// "i 0 199 45 0"
	index := db.indexs[key]

	buffer := make([]byte, index.size)

	db.pages[db.path+index.page].ReadAt(buffer, index.at)

	return string(buffer)
}

func (db *Database) ShowIndexs() {

	for k, v := range db.indexs {
		fmt.Println(k, v)
	}
	fmt.Println("len indexs : ", len(db.indexs))
}

// Open initialaze db pages
func Open(path string) *Database {

	db := &Database{}

	db.indexs = db.reIndex() //map[string][3]int64{}
	db.pages = make(map[string]*os.File)
	afile := "0" // active file
	db.path = path

	if db.path == "" {
		//path, _ = os.Getwd()
		db.path = "mok/"

		err := os.Mkdir(db.path, 0744)
		check("Mkdir ", err)

		db.afile = db.path + afile // active file

		file, err := os.OpenFile(db.afile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		check("when open file", err)

		fmt.Println("file active is : ", file.Name())
		db.pages[db.afile] = file

		// complet db initalaze

		return db
	}

	fmt.Println("read dir", db.path)

	err := os.Mkdir(db.path, 0744)
	check("Mkdir ", err)

	db.afile = db.path + afile // active file

	file, err := os.OpenFile(db.afile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	check("when open file", err)
	file.Close()

	files, err := os.ReadDir(db.path)
	check("ReadDir ", err)

	for k, f := range files {
		fmt.Println(f.Name())

		dpage := db.path + f.Name()

		file, err := os.OpenFile(dpage, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		check("", err)

		db.pages[fmt.Sprint(k)] = file
		fmt.Println("f.name:", f.Name())
		fmt.Println("k:", k)

		fmt.Println("data page : ", dpage)
		db.pages[dpage] = file

		db.afile = dpage
		fmt.Println("file active is : ", db.afile)
	}

	fmt.Println("afile : ", afile)
	fmt.Printf("pages :  %v\n", db.pages)

	return db
}

// rebuilds indexs
func (db *Database) reIndex() (indexs map[string]index) {
	// Read the entire file into a byte slice

	indexs = make(map[string]index)
	//indexs = make(map[string][3]int64)

	for f := range db.pages {
		fileContent, err := os.ReadFile(f)
		check("", err)

		// Split the byte slice into lines using the newline character as a delimiter
		lines := strings.Split(string(fileContent), "\n")

		// Process each line
		for _, line := range lines {
			if len(line) == 0 {
				return
			}
			if line[0] == 'i' {

				pos := strings.Fields(line)
				at, _ := strconv.Atoi(pos[2])
				size, _ := strconv.Atoi(pos[3])
				//page, _ := strconv.Atoi(pos[4])

				indexs[pos[1]] = index{at: int64(at), size: size}
			}
		}

	}
	fmt.Println(indexs)

	return indexs
}

// Close db
func (db *Database) Close() {
	for _, f := range db.pages {
		f.Close()
	}
}

// error
func check(hint string, err error) {
	if err != nil {
		fmt.Println(hint, err)
		//return
	}
}

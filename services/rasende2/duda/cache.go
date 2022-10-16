package duda

import (
	"errors"
	"log"
	"os"
	"strings"
)

type Cache struct {
	directory string
}

func NewCache(directory string) *Cache {
	if !strings.HasSuffix(directory, "/") {
		directory = directory + "/"
	}
	os.MkdirAll(directory, os.ModePerm)
	return &Cache{
		directory: directory,
	}
}

func (c *Cache) Get(key string) (string, bool) {
	filepath := c.directory + key
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("error reading cache file: %v", err)
		}
		return "", false
	}
	return string(bytes), true
}

func (c *Cache) Put(key string, value string) {
	filepath := c.directory + key
	filepathParts := strings.Split(filepath, "/")
	filepathDir := strings.Join(filepathParts[:len(filepathParts)-1], "/")
	os.MkdirAll(filepathDir, os.ModePerm)
	f, err := os.Create(filepath)
	if err != nil {
		log.Printf("error creating cache file: %v", err)
	}
	defer f.Close()
	_, err = f.WriteString(value)
	if err != nil {
		log.Printf("error writing cache file: %v", err)
	}
}

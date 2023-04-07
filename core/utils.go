package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// ComputeHash computes hash of the given raw message
func ComputeHash(raw []byte) ([]byte, error) {
	hash := sha256.New()
	hash.Write(raw)
	return hash.Sum(nil), nil
}

type Configuration struct {
	Id        uint8    `json:"id"`
	N         uint8    `json:"n"`
	Port      uint64   `json:"port"`
	Committee []uint64 `json:"committee"`
	BlockSize uint64   `json:"blocksize"`
}

func GetConfig(id int) *Configuration {
	configFile := fmt.Sprintf("../config/node%d.json", id)
	jsonFile, err := os.Open(configFile)
	if err != nil {
		panic(fmt.Sprint("os.Open: ", err))
	}
	defer jsonFile.Close()

	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(fmt.Sprint("ioutil.ReadAll: ", err))
	}
	var config Configuration
	json.Unmarshal([]byte(data), &config)
	return &config
}

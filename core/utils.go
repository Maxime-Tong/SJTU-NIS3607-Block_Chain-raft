package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
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

func afterBetween(min time.Duration, max time.Duration) <-chan time.Time {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	d, delta := min, (max - min)
	if delta > 0 {
		d += time.Duration(rand.Int63n(int64(delta)))
	}
	return time.After(d)
}

func Min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Username   string `json:"username"`
	ServerPort int    `json:"serverPort"`
}

func GetConfigFromFile() (config *Config, err error) {
	defaultFile := os.Getenv("HOME") + "/.hichat_conf.json"
	configFile, err := os.Open(defaultFile)
	if err != nil {
		return
	}
	defer configFile.Close()
	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		return
	}
	var conf Config
	json.Unmarshal(byteValue, &conf)
	config = &conf
	return
}

package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	BindAddr string   `json:"bind_addr"`
	BindPort string   `json:"bind_port"`
	LogFile  string   `json:"log_file"`
	Rooms    []string `json:"rooms"`
}

var Config *Configuration

func ReadConfiguration(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &Config)
}

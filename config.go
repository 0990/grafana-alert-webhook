package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct{
	Listen string `json:"listen"`
	CorpID string `json:"corpid"`
	CorpSecret string `json:"corpsecret"`
	AgentId int `json:"agentid"`
	ToUserDefault string `json:"touserdefault"`
}

func readOrCreateCfg(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := createCfg(path)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return readCfg(path)
}

func readCfg(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func createCfg(path string) error {
	c, _ := json.MarshalIndent(Config{}, "", "    ")
	return ioutil.WriteFile(path, c, 0644)
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Listen   string         `json:"listen"`
	WeCom    WeComConfig    `json:"wecom,omitempty"`
	XiaoShan XiaoShanConfig `json:"xiaoshan,omitempty"`
}

type WeComConfig struct {
	Enable        bool   `json:"enable"`
	CorpID        string `json:"corpid"`
	CorpSecret    string `json:"corpsecret"`
	AgentID       int    `json:"agentid"`
	ToUserDefault string `json:"touserdefault,omitempty"`
	ToTagDefault  string `json:"totagdefault,omitempty"`
}

type XiaoShanConfig struct {
	Enable        bool     `json:"enable"`
	URL           string   `json:"url,omitempty"`
	AccessToken   string   `json:"accesstoken"`
	Secret        string   `json:"secret"`
	MsgType       string   `json:"msgtype,omitempty"`
	AtAll         bool     `json:"atall,omitempty"`
	AtMobiles     []string `json:"atmobiles,omitempty"`
	AtUserIDs     []string `json:"atuserids,omitempty"`
	MaxTextLength int      `json:"maxtextlength,omitempty"`
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

	conf.normalize()
	return &conf, nil
}

func createCfg(path string) error {
	c, _ := json.MarshalIndent(defaultCfg(), "", "    ")
	return ioutil.WriteFile(path, c, 0644)
}

func defaultCfg() Config {
	return Config{
		Listen: ":1111",
		WeCom:  WeComConfig{},
	}
}

func (c *Config) normalize() {
	if c.Listen == "" {
		c.Listen = ":1111"
	}
}

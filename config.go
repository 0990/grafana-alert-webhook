package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Listen  string         `json:"listen"`
	Pushers []PusherConfig `json:"pushers,omitempty"`

	CorpID        string `json:"corpid,omitempty"`
	CorpSecret    string `json:"corpsecret,omitempty"`
	AgentID       int    `json:"agentid,omitempty"`
	ToUserDefault string `json:"touserdefault,omitempty"`
}

type PusherConfig struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Enabled *bool           `json:"enabled,omitempty"`
	Config  json.RawMessage `json:"config,omitempty"`
	WeCom   WeComConfig     `json:"wecom,omitempty"`
}

type WeComConfig struct {
	CorpID        string `json:"corpid"`
	CorpSecret    string `json:"corpsecret"`
	AgentID       int    `json:"agentid"`
	ToUserDefault string `json:"touserdefault,omitempty"`
	ToTagDefault  string `json:"totagdefault,omitempty"`
}

func (p PusherConfig) IsEnabled() bool {
	return p.Enabled == nil || *p.Enabled
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
	enabled := true
	return Config{
		Listen: ":1111",
		Pushers: []PusherConfig{
			{
				Name:    "default-wecom",
				Type:    "wecom",
				Enabled: &enabled,
				Config:  rawConfig(WeComConfig{}),
			},
		},
	}
}

func (c *Config) normalize() {
	if c.Listen == "" {
		c.Listen = ":1111"
	}
	if len(c.Pushers) > 0 {
		return
	}
	if c.CorpID == "" && c.CorpSecret == "" && c.AgentID == 0 && c.ToUserDefault == "" {
		return
	}

	enabled := true
	c.Pushers = []PusherConfig{
		{
			Name:    "default-wecom",
			Type:    "wecom",
			Enabled: &enabled,
			Config: rawConfig(WeComConfig{
				CorpID:        c.CorpID,
				CorpSecret:    c.CorpSecret,
				AgentID:       c.AgentID,
				ToUserDefault: c.ToUserDefault,
			}),
		},
	}
}

func rawConfig(value interface{}) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}

package streamer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Secret   string `json:"secret"`
	Username string `json:"username"`
}

func NewCfg() Config {
	return Config{}
}

func ReadCfg(path string) (Config, error) {
	cfg := NewCfg()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func WriteCfg(path string, cfg Config) error {
	jsonString, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jsonString, 0644)
}

func UpdateCfg(path, key string, value string) (Config, error) {
	cfg, err := ReadCfg(path)
	if err != nil {
		return cfg, err
	}
	switch key {
	case "Username":
		cfg.Username = value
	case "Secret":
		cfg.Secret = value
	default:
		return cfg, fmt.Errorf("Unknown key: %s", key)
	}

	err = WriteCfg(path, cfg)
	return cfg, err
}

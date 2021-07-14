package streamer

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var CONFIG_PATH = os.ExpandEnv("$HOME/.tstream.conf")

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

func GetSecret(configPath string) string {
	cfg, err := ReadCfg(CONFIG_PATH)
	var secret string

	// gen a new one if not existed
	if err != nil {
		cfg = NewCfg()
		cfg.Secret = GenSecret("tstream")
		WriteCfg(CONFIG_PATH, cfg)
	} else {
		secret = cfg.Secret
	}
	return secret
}

func GenSecret(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

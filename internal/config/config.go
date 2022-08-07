package config

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
)

type Config struct {
	Nats NatsConfig `json:"nats"`
	DB   DbConfig   `json:"db"`
}

type NatsConfig struct {
	Port      string `json:"port"`
	ClienID   string `json:"client_id"`
	ClusterID string `json:"cluster_id"`
}

type DbConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	Port     string `json:"port"`
}

func GetConfig(isTest bool) *Config {
	cfgName := "config"
	if isTest {
		cfgName += "_test"
	}
	cfgName += ".json"

	_, b, _, _ := runtime.Caller(0)
	path := path.Join(path.Dir(b), cfgName)

	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}

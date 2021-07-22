package src

import (
	"github.com/BurntSushi/toml"
)

type SimConfig struct {
	NumWorkers      int
	MaxUsersPerTick int
	SitemapURL      string

	Producer ProducerConfig
}

func NewSimConfigFromFile(file string) (*SimConfig, error) {
	var config SimConfig
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

type ApiConfig struct {
	SingleStore SingleStoreConfig
}

func NewApiConfigFromFile(file string) (*ApiConfig, error) {
	var config ApiConfig
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

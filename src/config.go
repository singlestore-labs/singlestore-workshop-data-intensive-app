package src

import (
	"github.com/BurntSushi/toml"
)

type SimConfig struct {
	NumWorkers int
	Producer   ProducerConfig

	// if you need to connect directly to SingleStore in your simulator
	// define that here
	// SingleStore SingleStoreConfig
}

func NewSimConfigFromFile(file string) (*SimConfig, error) {
	var config SimConfig
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

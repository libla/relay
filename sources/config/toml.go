package config

import (
	"relay"

	"github.com/BurntSushi/toml"
)

func FromToml(content []byte) (relay.Config, error) {
	var root map[string]any
	_, err := toml.Decode(string(content), &root)
	if err != nil {
		return nil, err
	}
	return FromMap(root), nil
}

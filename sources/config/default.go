package config

import "relay"

func Default(ext string) func([]byte) (relay.Config, error) {
	switch ext {
	case "toml":
		return FromToml
	}
	return nil
}

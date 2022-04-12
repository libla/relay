package relay

type Module interface {
	Load(config Config) error
	Unload() error
}

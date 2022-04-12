package relay

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var Application, cancel = context.WithCancel(context.Background())
var Void struct{}

var loops sync.WaitGroup
var errloop = make(chan error)
var factories = make(map[string]func() Module)

func Register(name string, factory func() Module) {
	factories[name] = factory
}

func LoadArgs(creators ...func(string) func([]byte) (Config, error)) (Config, error) {
	var configs []Config
	for _, arg := range os.Args[1:] {
		ext := filepath.Ext(arg)
		if len(ext) != 0 {
			ext = ext[1:]
		}
		ext = strings.ToLower(ext)
		for _, creator := range creators {
			load := creator(ext)
			if load != nil {
				content, err := ioutil.ReadFile(arg)
				if err != nil {
					return nil, err
				}
				config, err := load(content)
				if err != nil {
					return nil, err
				}
				configs = append(configs, config)
				break
			}
		}
	}
	length := len(configs)
	if length == 0 {
		return EmptyConfig(), nil
	}
	if length == 1 {
		return configs[0], nil
	}
	return ConfigCombine(configs[0], configs[1:]...), nil
}

func Bootstrap(config Config) error {
	module := config.GetConfig("module")
	if module == nil {
		module = EmptyConfig()
	}
	common := ConfigSkip(config, "module")
	if common.Empty() {
		common = EmptyConfig()
	}
	modules, err := loadModules(common, module)
	if err != nil {
		return err
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()
	go func() {
		loops.Wait()
		errloop <- nil
	}()
	err = <-errloop
	if err != nil {
		cancel()
		go func() {
			for {
				<-errloop
			}
		}()
		loops.Wait()
		unloadModules(modules)
		return err
	}
	return unloadModules(modules)
}

func Shutdown() {
	cancel()
}

func loadModules(common, module Config) ([]Module, error) {
	ch := make(chan error)
	wait := sync.WaitGroup{}
	var guard sync.Mutex
	var modules []Module
	keys := module.Keys()
	for {
		key := keys.Next()
		if key == nil {
			break
		}
		config := module.GetConfig(*key)
		if config != nil {
			if common != EmptyConfig() {
				config = ConfigCombine(config, common)
			}
			if factory, ok := factories[*key]; ok {
				module := factory()
				wait.Add(1)
				go func() {
					defer wait.Done()
					err := module.Load(config)
					if err != nil {
						ch <- err
					} else {
						guard.Lock()
						defer guard.Unlock()
						modules = append(modules, module)
					}
				}()
			}
		}
	}
	go func() {
		wait.Wait()
		ch <- nil
	}()
	err := <-ch
	if err != nil {
		go func() {
			for {
				<-ch
			}
		}()
		wait.Wait()
		unloadModules(modules)
		return nil, err
	}
	return modules, nil
}

func unloadModules(modules []Module) error {
	ch := make(chan error, len(modules))
	wait := sync.WaitGroup{}
	for i := 0; i < len(modules); i++ {
		module := modules[i]
		wait.Add(i)
		go func() {
			defer wait.Done()
			err := module.Unload()
			if err != nil {
				ch <- err
			}
		}()
	}
	go func() {
		wait.Wait()
		ch <- nil
	}()
	return <-ch
}

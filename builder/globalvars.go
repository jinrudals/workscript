package builder

import (
	"fmt"
	"strings"
	"sync"
)

type GlobalVars struct {
	mu   sync.RWMutex
	mmap map[string]string
	run  map[string]bool
}

var (
	instance *GlobalVars
	once     sync.Once
)

func GetInstance() *GlobalVars {
	once.Do(func() {
		instance = &GlobalVars{
			mmap: make(map[string]string),
			run:  make(map[string]bool),
		}
	})
	return instance
}

// Set accepts mixed-type values: strings for mmap, bools for run.
func (g *GlobalVars) Set(items map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for key, value := range items {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid global key format: %s", key)
		}
		stageName, envName := parts[0], parts[1]

		switch stageName {
		case "run":
			boolVal, ok := value.(bool)
			if !ok {
				return fmt.Errorf("expected boolean for run:%s, got %T", envName, value)
			}
			g.run[envName] = boolVal

		default:
			strVal, ok := value.(string)
			if !ok {
				return fmt.Errorf("expected string for %s:%s, got %T", stageName, envName, value)
			}
			fullKey := fmt.Sprintf("%s:%s", stageName, envName)
			g.mmap[fullKey] = strVal
		}
	}
	return nil
}

func (g *GlobalVars) Get(key string) (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	val, ok := g.mmap[key]
	return val, ok
}

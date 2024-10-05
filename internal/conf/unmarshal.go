package conf

import (
	"os"
	"path/filepath"

	"github.com/ixugo/goweb/pkg/system"
	"github.com/pelletier/go-toml/v2"
)

func SetupConfig(v any) error {
	path := filepath.Join(system.GetCWD(), "configs", "config.toml")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, v)
}

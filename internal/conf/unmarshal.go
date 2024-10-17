package conf

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

func SetupConfig(v any, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, v)
}

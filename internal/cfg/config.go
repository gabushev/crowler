package cfg

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Parallelism         int      `yaml:"parallelism"`
	AcceptableMimeTypes []string `yaml:"acceptable_mime_types"`
	DatabaseFile        string   `yaml:"database_file"`
	ApiAddr             string   `yaml:"api_addr"`
	DownloadsDir        string   `yaml:"downloads_dir"`
}

func NewConfig(configPath string) (*Config, error) {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

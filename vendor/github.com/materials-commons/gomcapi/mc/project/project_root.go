package project

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func FindRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		mcDir := filepath.Join(currentDir, ".mc")
		if _, err := os.Stat(mcDir); !os.IsNotExist(err) {
			return mcDir, nil
		}
		currentDir = filepath.Dir(currentDir)
		if currentDir == "/" {
			return "", os.ErrNotExist
		}
	}
}

type Config struct {
	ID string `json:"id"`
}

func FindConfig() (*Config, error) {
	var (
		mcDir  string
		err    error
		config Config
	)

	if mcDir, err = FindRoot(); err != nil {
		var b []byte
		if b, err = ioutil.ReadFile(filepath.Join(mcDir, "project.json")); err != nil {
			err = json.Unmarshal(b, &config)
		}
	}

	return &config, err
}

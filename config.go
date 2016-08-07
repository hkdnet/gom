package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
)

type config struct {
	token string
}

func newConfig() (config, error) {
	token, err := readGemConfig()
	if err != nil {
		return config{}, errors.Wrap(err, "failed to load config")
	}
	return config{token: token}, nil
}

func readGemConfig() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadFile(filepath.Join(home, ".gem", "credentials"))
	if err != nil {
		return "", err
	}
	tmp := struct {
		APIKey string `yaml:":rubygems_api_key"`
	}{}
	err = yaml.Unmarshal(b, &tmp)
	return tmp.APIKey, err
}

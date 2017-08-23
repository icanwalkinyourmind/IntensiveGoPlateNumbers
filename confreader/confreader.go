package confreader

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Server  string `yaml:"server"`
	Port    string `yaml:"port"`
	Workers int    `yaml:"workers"`
}

func (c *Config) ReadConfig(fName string, conf interface{}) error {
	file, err := os.Open(fName)
	if err != nil {
		return fmt.Errorf("can't open YAML file %q: %s", fName, err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read yaml file %q: %s", fName, err)
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return fmt.Errorf("can't write YAML data into file %q: %s", fName, err)
	}

	return nil
}

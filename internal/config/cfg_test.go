package config

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestWriteConfig(t *testing.T) {
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	viper.SetConfigType("yaml")

	err = viper.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}

	err = viper.WriteConfigAs("../../default_config.yaml")
	if err != nil {
		t.Fatal(err)
	}
}

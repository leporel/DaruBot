package cmd

import (
	"DaruBot/internal/config"
	"github.com/sanity-io/litter"
	"github.com/spf13/viper"
	"testing"
)

func TestReadConfig(t *testing.T) {
	viper.SetConfigName("default_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../")

	err := viper.ReadInConfig()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.GetDefaultConfig()

	err = viper.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	pretty := litter.Options{StripPackageNames: true, HidePrivateFields: true}

	t.Logf("%s", pretty.Sdump(cfg))
}

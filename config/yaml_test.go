package config

import (
	"testing"
)

func TestYamlConfig(t *testing.T) {
	yamlConfig := NewYamlConfig("../log/log.yaml")
	t.Logf("%+v", yamlConfig)
	res, _ := yamlConfig.Maps("log::loggers")
	t.Logf("%+v", res)
}

package config

import (
	"io/ioutil"
	"strings"

	"github.com/smallfish/simpleyaml"
)

type YamlConfig struct {
	*simpleyaml.Yaml
}

func NewYamlConfig(filename string) Configer {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	yaml, err := simpleyaml.NewYaml(content)
	if err != nil {
		panic(err)
	}
	yamlConfig := &YamlConfig{
		Yaml: yaml,
	}
	return yamlConfig
}

func (config *YamlConfig) getNode(nodeName string) *simpleyaml.Yaml {
	nodes := strings.Split(nodeName, "::")
	if len(nodes) == 1 {
		nodes = append([]string{"default"}, nodes...)
	}
	return config.GetPath(nodes...)
}

func (config *YamlConfig) String(nodeName string) string {
	node := config.getNode(nodeName)
	res, err := node.String()
	if err != nil {
		return ""
	}
	return res
}

func (config *YamlConfig) Strings(nodeName string) []string {
	node := config.getNode(nodeName)
	res, err := node.Array()
	if err != nil {
		return nil
	}
	result := make([]string, 0, len(res))
	for _, item := range res {
		result = append(result, item.(string))
	}
	return result
}

func (config *YamlConfig) Int(nodeName string) (int, error) {
	node := config.getNode(nodeName)
	return node.Int()
}

func (config *YamlConfig) Ints(nodeName string) ([]int, error) {
	node := config.getNode(nodeName)
	res, err := node.Array()
	if err != nil {
		return nil, err
	}

	result := make([]int, 0, len(res))
	for _, item := range res {
		result = append(result, item.(int))
	}
	return result, nil
}

func (config *YamlConfig) Float(nodeName string) (float64, error) {
	return 0, nil
}

func (config *YamlConfig) Floats(nodeName string) (result []float64, err error) {
	return
}

func (config *YamlConfig) Bool(nodeName string) (bool, error) {
	return config.getNode(nodeName).Bool()
}

func (config *YamlConfig) Bools(nodeName string) ([]bool, error) {
	node := config.getNode(nodeName)
	res, err := node.Array()
	if err != nil {
		return nil, err
	}

	result := make([]bool, 0, len(res))
	for _, item := range res {
		result = append(result, item.(bool))
	}
	return result, nil
}

func (config *YamlConfig) Map(nodeName string) (map[string]interface{}, error) {
	res, err := config.getNode(nodeName).Map()
	if err != nil {
		return nil, err
	}
	return mapConvert(res), nil
}



func (config *YamlConfig) Maps(nodeName string) ([]map[string]interface{}, error) {
	res, err := config.getNode(nodeName).Array()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(res))
	for _, item := range res {
		result = append(result, mapConvert(
			item.(map[interface{}]interface{})))
	}
	return result, nil
}

func mapConvert(in map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		if mapValue, ok := value.(map[interface{}]interface{}); ok {
			out[key.(string)] = mapConvert(mapValue)
		} else {
			out[key.(string)] = value
		}
	}
	return out
}

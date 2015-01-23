package parser

import (
	"io"
	"os"

	"github.com/cloudfoundry-incubator/cli-plugin-repo/models"
	"github.com/cloudfoundry/cli/generic"
	"github.com/fraenkel/candiedyaml"
)

type YamlParser interface {
	Parse() (models.Plugins, error)
}

type yamlParser struct {
	filePath     string
	logger       io.Writer
	pluginsModel models.PluginModel
}

func NewYamlParser(filePath string, logger io.Writer, pluginsModel models.PluginModel) YamlParser {
	return yamlParser{
		filePath:     filePath,
		logger:       logger,
		pluginsModel: pluginsModel,
	}
}

func (p yamlParser) Parse() (models.Plugins, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		p.logger.Write([]byte("File does not exist:" + err.Error()))
		return models.Plugins{}, err
	}

	document := new(interface{})
	decoder := candiedyaml.NewDecoder(file)
	err = decoder.Decode(document)

	if err != nil {
		p.logger.Write([]byte("Failed to decode document:" + err.Error()))
		return models.Plugins{}, err
	}

	output, _ := expandProperties(*document)

	p.pluginsModel.PopulateModel(output)

	return p.pluginsModel.PluginsModel(), nil
}

func expandProperties(input interface{}) (output interface{}, errs []error) {
	switch input := input.(type) {
	case string:
		output = input
	case []interface{}:
		outputSlice := make([]interface{}, len(input))
		for index, item := range input {
			itemOutput, itemErrs := expandProperties(item)
			outputSlice[index] = itemOutput
			errs = append(errs, itemErrs...)
		}
		output = outputSlice
	case map[interface{}]interface{}:
		outputMap := make(map[interface{}]interface{})
		for key, value := range input {
			itemOutput, itemErrs := expandProperties(value)
			outputMap[key] = itemOutput
			errs = append(errs, itemErrs...)
		}
		output = outputMap
	case generic.Map:
		outputMap := generic.NewMap()
		generic.Each(input, func(key, value interface{}) {
			itemOutput, itemErrs := expandProperties(value)
			outputMap.Set(key, itemOutput)
			errs = append(errs, itemErrs...)
		})
		output = outputMap
	default:
		output = input
	}

	return
}

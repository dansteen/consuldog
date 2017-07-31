package datadog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/dansteen/consuldog/services"
	"github.com/spf13/viper"
)

// WriteConfig will write out monitoring files for datadog based on the information provided in the services we have stored
// It will always write all config files it knows about
func WriteConfig(allServices services.Services) {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// a place to store all of our config Objects once they are populated
	configObjects := make(map[string]CheckConf)
	tmpBuf := new(bytes.Buffer)
	// get the templates we will need
	templates := getConfTemplates(allServices)

	// run through our services by type and generate datdog config files
	for datadogType, services := range allServices.ByType {
		// create our aggregate config for this type
		typeConfig := CheckConf{
			InitConfig: make([]interface{}, 0),
			Instances:  make([]interface{}, 0),
		}

		// run through each service of this type
		for _, service := range services {
			// and instantiate our template if it exists
			if ourTemplate, found := templates[service.ConfigTemplate]; found {
				err := ourTemplate.Execute(tmpBuf, service)
				if err != nil {
					logger.Println(err)
					logger.Printf("Could not execute template %s for service %s. Skipping.\n", service.ConfigTemplate, service.Service)
					continue
				}
				// if we did not find the template move on
			} else {
				logger.Printf("Could not find template %s for service %s. Skipping.\n", service.ConfigTemplate, service.Service)
				continue
			}

			// once we have the template, unMarshal the yaml
			var config CheckConf
			err := yaml.Unmarshal(tmpBuf.Bytes(), &config)
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not convert template %s to object for service %s. Skipping.\n", service.ConfigTemplate, service.Service)
				continue
			}

			// once we've gotten to this point things look good so we add this config into our final config
			for _, initConfig := range config.InitConfig {
				typeConfig.InitConfig = append(typeConfig.InitConfig, initConfig)
			}
			for _, instance := range config.Instances {
				typeConfig.Instances = append(typeConfig.Instances, instance)
			}
		}

		// once we are done, add this typeConfig to our list
		configObjects[datadogType] = typeConfig
	}

	// after we are done generating all of our configs, we write them out to config files
	for datadogType, config := range configObjects {
		fileBytes, err := yaml.Marshal(config)
		if err != nil {
			logger.Println(err)
			logger.Printf("Could not convert %s config to yaml file. Skipping.\n", datadogType)
			continue
		}
		// put our datadog check filename together
		ddFilePath := path.Join(viper.GetString("datadogFolder"), "conf.d", fmt.Sprintf("%s.yaml", datadogType))

		err = ioutil.WriteFile(ddFilePath, fileBytes, 0644)
		if err != nil {
			logger.Println(err)
			logger.Printf("Could not write file %s. Skipping.\n", ddFilePath)
			continue
		}
	}
}

// getConfigTemplates will generate a map of templates keyed on service.ConfigTemplate for all templates that are required by allServices
// templates must be valid yaml in the correct datadogFormat or it will be skipped
func getConfTemplates(allServices services.Services) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	// create our logger
	logger := log.New(os.Stderr, log.Prefix(), 0)
	for _, service := range allServices.Services {
		// get the full path to the template
		templatePath := path.Join(viper.GetString("templateFolder"), service.ConfigTemplate)
		// read in the raw template file
		rawTemplate, err := ioutil.ReadFile(templatePath)
		if err != nil {
			logger.Println(err)
			logger.Printf("Could not load template for %s. Skipping.\n", service.ConfigTemplate)
			continue
		}

		// make sure its valid YAML and conforms to the structrue we need for datadog
		var config CheckConf
		err = yaml.Unmarshal(rawTemplate, &config)
		if err != nil {
			logger.Println(err)
			logger.Printf("%s is not valid YAML (or does not conform to our required structure) for %s. Please ensure its formatted correctly.  Skipping.\n", templatePath, service.ConfigTemplate)
			continue
		}

		// once we have loaded it and are sure it's valid YAML we can turn it into a template
		tmpl, err := template.New(service.ConfigTemplate).Parse(string(rawTemplate))
		if err != nil {
			logger.Println(err)
			logger.Printf("Could not load template for %s. Skipping.\n", service.ConfigTemplate)
			continue
		}
		templates[service.ConfigTemplate] = tmpl

	}
	return templates
}

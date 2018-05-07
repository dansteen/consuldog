package datadog

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/dansteen/consuldog/services"
	consul "github.com/hashicorp/consul/api"
	getter "github.com/hashicorp/go-getter"
	"github.com/spf13/viper"
)

// WriteConfig will write out monitoring files for datadog based on the information provided in the services we have stored
// It will always write all config files it knows about
func WriteConfig(allServices services.Services) {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// a place to store all of our config Objects once they are populated
	configObjects := make(map[string]CheckConf)
	// get the templates we will need
	templates := getConfTemplates(allServices)

	// run through our services by type and generate datdog config files
	for datadogType, monitors := range allServices.MonitorByType {
		// create our aggregate config for this type
		typeConfig := CheckConf{
			InitConfig: make(map[string]interface{}),
			Instances:  make([]interface{}, 0),
		}

		// run through each service of this type
		for _, monitor := range monitors {
			tmpBuf := new(bytes.Buffer)
			// and instantiate our template if it exists
			if ourTemplate, found := templates[monitor.ConfigTemplate]; found {
				err := ourTemplate.Execute(tmpBuf, monitor.Service)
				if err != nil {
					logger.Println(err)
					logger.Printf("Could not execute template %s for service %s. Skipping.\n", monitor.ConfigTemplate, monitor.Service.Service)
					continue
				}
				// if we did not find the template move on
			} else {
				logger.Printf("Could not find template %s for service %s. Skipping.\n", monitor.ConfigTemplate, monitor.Service.Service)
				continue
			}

			// once we have the template, unMarshal the yaml
			var config CheckConf
			err := yaml.Unmarshal(tmpBuf.Bytes(), &config)
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not convert template %s to object for service %s. Skipping.\n", monitor.ConfigTemplate, monitor.Service.Service)
				continue
			}

			// once we've gotten to this point things look good so we add this config into our final config
			for initConfName, initConfValue := range config.InitConfig {
				typeConfig.InitConfig[initConfName] = initConfValue
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
	// variables to populate below
	//var templatePath string
	//var dudService services.Service
	for _, service := range allServices.Services {
		// for each monitor in our service
		for _, monitor := range service.Monitors {
			// first generate a temp filename
			b := make([]byte, 16)
			_, err := rand.Read(b)
			if err != nil {
				fmt.Println(err)
				logger.Printf("Warning: Could not get random string %s. Skipping.\n", monitor.ConfigTemplate)
				continue
			}
			randomString := base64.StdEncoding.EncodeToString(b)
			templatePath := path.Join(viper.GetString("tempFolder"), fmt.Sprintf("%s.%s", path.Base(monitor.ConfigTemplate), randomString))
			// then download the template file from the url provided
			err = getter.GetFile(templatePath, monitor.ConfigTemplate)
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not get template for %s. Skipping.\n", monitor.ConfigTemplate)
				continue
			}

			// read in the raw template file
			rawTemplate, err := ioutil.ReadFile(templatePath)
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not load template for %s. Skipping.\n", monitor.ConfigTemplate)
				continue
			}
			// if we can at least read the file we remove it.
			err = os.Remove(templatePath)
			if err != nil {
				logger.Println(err)
				logger.Printf("Warning: Could not remove temp file %s.\n", templatePath)
			}

			// turn our raw template string into a template object
			tmpl, err := template.New(monitor.ConfigTemplate).Parse(string(rawTemplate))
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not create template for %s. Skipping.\n", monitor.ConfigTemplate)
				continue
			}

			// YAML doesn't like {{ at the start of a scalar.  Unfortunately, this is common in our templates.  Fortunately, in usage, we de-template prior to actually UnMarshaling the template so here, when testing it we dud out the values first as well.
			dudService := services.Service{
				AgentService: consul.AgentService{
					Address:     "127.0.0.1",
					CreateIndex: 123456789,
					ModifyIndex: 123456789,
					ID:          "test-service-ID",
					Port:        9999,
					Service:     "test-service",
					Tags:        []string{"tag1", "tag2"},
				},
				Monitors: []services.Monitor{
					{
						ConfigTemplate: monitor.ConfigTemplate,
						DatadogType:    monitor.DatadogType,
					},
				},
				Node: "test-node",
			}
			// instantiate our dud
			dudInstance := new(bytes.Buffer)
			err = tmpl.Execute(dudInstance, dudService)
			if err != nil {
				logger.Println(err)
				logger.Printf("Could not execute template %s. Skipping.\n", monitor.ConfigTemplate)
				continue
			}

			// once we have an instantiated template make sure its valid YAML and conforms to the structrue we need for datadog
			var config CheckConf
			err = yaml.Unmarshal(dudInstance.Bytes(), &config)
			if err != nil {
				logger.Println(err)
				logger.Printf("%s is not valid YAML (or does not conform to our required structure) for %s. Please ensure its formatted correctly.  Skipping.\n", templatePath, monitor.ConfigTemplate)
				continue
			}

			// once we have the template and have verified its validity, we save it to our template store
			templates[monitor.ConfigTemplate] = tmpl
		}
	}
	return templates
}

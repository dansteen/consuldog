package services

import (
	consul "github.com/hashicorp/consul/api"
)

// Service contains details of services for a particular node, as well as the templates to use for that service
type Service struct {
	consul.AgentService
	ConfigTemplate string
	DatadogType    string
	Node           string
}

type NodeServices struct {
	Node     string
	Services []Service
}

// Services stores services per node, and handles writing them out to actual config files
type Services struct {
	Services map[string]*Service
	ByNode   map[string][]*Service
	ByType   map[string][]*Service
}

// NewServiceConfig will generate a new ServiceConfig object to populate with services
func NewServices() Services {
	return Services{
		Services: make(map[string]*Service),
		ByNode:   make(map[string][]*Service),
		ByType:   make(map[string][]*Service),
	}
}

// Add adds a new NodeService to our list of services, and overwrites all previous services for that node
func (services *Services) Add(newService Service) {
	services.Services[newService.ID] = &newService
	services.ByNode[newService.Node] = append(services.ByNode[newService.Node], &newService)
	services.ByType[newService.DatadogType] = append(services.ByType[newService.DatadogType], &newService)
}

// ClearNode will remove all services from a specific node
func (services *Services) ClearNode(nodeName string) {
	// run through our services
	for _, service := range services.ByNode[nodeName] {
		// remove them from ByType
		// first find it
		var foundIndex int
		for index, serviceByType := range services.ByType[service.DatadogType] {
			if serviceByType == service {
				foundIndex = index
				break
			}
		}
		// then remove it
		// this deletes it but does not preserve the order of the services (which is fine)
		// got this from here: https://github.com/golang/go/wiki/SliceTricks as it does not result in memory leaks
		services.ByType[service.DatadogType][foundIndex] = services.ByType[service.DatadogType][len(services.ByType[service.DatadogType])-1]
		services.ByType[service.DatadogType][len(services.ByType[service.DatadogType])-1] = nil
		services.ByType[service.DatadogType] = services.ByType[service.DatadogType][:len(services.ByType[service.DatadogType])-1]
		// then delete it from our list of services
		delete(services.Services, service.ID)
	}
	// once we have removed all of our services, we remove them from ByNode as well
	delete(services.ByNode, nodeName)
}

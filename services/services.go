package services

import (
	consul "github.com/hashicorp/consul/api"
)

// Monitor contains information about a particular monitor for a service
type Monitor struct {
	ConfigTemplate string
	DatadogType    string
	Service        *Service
}

// Service contains details of services for a particular node, as well as the templates to use for that service
type Service struct {
	consul.AgentService
	Monitors []Monitor
	Node     string
}

type NodeServices struct {
	Node     string
	Services []Service
}

// Services stores services per node, and handles writing them out to actual config files
type Services struct {
	Services      map[string]*Service
	ByNode        map[string][]*Service
	MonitorByType map[string][]*Monitor
}

// NewServiceConfig will generate a new ServiceConfig object to populate with services
func NewServices() Services {
	return Services{
		Services:      make(map[string]*Service),
		ByNode:        make(map[string][]*Service),
		MonitorByType: make(map[string][]*Monitor),
	}
}

// Add adds a new NodeService to our list of services, and overwrites all previous services for that node
func (services *Services) Add(newService Service) {
	services.Services[newService.ID] = &newService
	services.ByNode[newService.Node] = append(services.ByNode[newService.Node], &newService)

	// for each monitor add an entry into MonitorByType so we can pull them out later
	for _, monitor := range newService.Monitors {
		services.MonitorByType[monitor.DatadogType] = append(services.MonitorByType[monitor.DatadogType], &Monitor{
			ConfigTemplate: monitor.ConfigTemplate,
			DatadogType:    monitor.DatadogType,
			Service:        monitor.Service,
		})
	}
}

// ClearNode will remove all services from a specific node
func (services *Services) ClearNode(nodeName string) {
	// run through our services
	for _, service := range services.ByNode[nodeName] {
		for _, monitor := range service.Monitors {
			// remove the monitors from MonitorByType
			// first find it
			var foundIndex int
			for index, monitorByType := range services.MonitorByType[monitor.DatadogType] {
				if *monitorByType == monitor {
					foundIndex = index
					break
				}
			}
			// then remove it
			// this deletes it but does not preserve the order of the services (which is fine)
			// got this from here: https://github.com/golang/go/wiki/SliceTricks as it does not result in memory leaks
			services.MonitorByType[monitor.DatadogType][foundIndex] = services.MonitorByType[monitor.DatadogType][len(services.MonitorByType[monitor.DatadogType])-1]
			services.MonitorByType[monitor.DatadogType][len(services.MonitorByType[monitor.DatadogType])-1] = nil
			services.MonitorByType[monitor.DatadogType] = services.MonitorByType[monitor.DatadogType][:len(services.MonitorByType[monitor.DatadogType])-1]
		}
		// then delete our service from our list of services
		delete(services.Services, service.ID)
	}
	// once we have removed all of our services, we remove them from ByNode as well
	delete(services.ByNode, nodeName)
}

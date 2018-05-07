package communicator

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/dansteen/consuldog/services"
	consul "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

// ConsulClient stores information about our communication with consul
type ConsulClient struct {
	client *consul.Client
}

// NewConsulClient will generate a new connection to consul
func NewConsulClient(consulAddress string) ConsulClient {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// configure our consul client
	config := consul.Config{
		Address: consulAddress,
	}
	consulClient, err := consul.NewClient(&config)
	if err != nil {
		logger.Fatal(err)
	}
	return ConsulClient{
		client: consulClient,
	}
}

// MonitorNode will monitor consul for changes in a node and, on changes, send back a list of service for that node
// that match our prefix
func (consulClient *ConsulClient) MonitorNode(node string, serviceOut chan<- services.NodeServices, cont <-chan bool) {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// grab our catalog connection
	catalog := consulClient.client.Catalog()
	// we want to return right away the first time so we get an initial set of services
	lastIndex := uint64(0)
	// keep going until we are told to stop
	for {
		select {
		case <-cont:
			return
		default:
			node, meta, err := catalog.Node(node, &consul.QueryOptions{
				AllowStale: true,
				WaitIndex:  lastIndex,
			})
			// if we get an error we wait and then try again
			if err != nil {
				logger.Println(err)
				time.Sleep(5 * time.Second)
			} else {
				if lastIndex != meta.LastIndex {
					lastIndex = meta.LastIndex
					// create our NodeServices object
					foundServices := services.NodeServices{
						Node:     node.Node.Node,
						Services: make([]services.Service, 0),
					}

					// create a list of services to be monitored
					for _, service := range node.Services {
						// generate our service
						var newService services.Service
						newService = services.Service{
							Monitors:     make([]services.Monitor, 0),
							AgentService: *service,
							Node:         node.Node.Node,
						}

						// grab our tags that have our prefix
						for _, tag := range service.Tags {
							if strings.HasPrefix(tag, viper.GetString("prefix")) {
								// parse our values
								values := strings.SplitN(strings.TrimPrefix(tag, viper.GetString("prefix")), " ", 2)
								// and create monitors for them
								newService.Monitors = append(newService.Monitors, services.Monitor{
									ConfigTemplate: values[0],
									DatadogType:    values[1],
									Service:        &newService,
								})
							}
						}
						// if we found monitors, add that service to our list
						if len(newService.Monitors) > 0 {
							foundServices.Services = append(foundServices.Services, newService)
						}
					}
					// we always return if there was an updat since we need to know if services were removed
					serviceOut <- foundServices
				}
			}
		}
	}
}

// GetNodeName will get the node name of the consul agent we have connected to
func (consulClient *ConsulClient) GetNodeName() string {
	// log errors to stderr
	logger := log.New(os.Stderr, log.Prefix(), 0)
	// we keep trying to connect
	for {
		nodeName, err := consulClient.client.Agent().NodeName()
		if err != nil {
			logger.Println(err)
		} else {
			return nodeName
		}
		time.Sleep(5 * time.Second)
	}
}

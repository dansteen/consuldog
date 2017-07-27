package communicator

import (
	"log"
	"strings"
	"time"

	"github.com/dansteen/consuldog/services"
	consul "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

type ConsulClient struct {
	client *consul.Client
}

// NewConsulClient will generate a new connection to consul
func NewConsulClient(consulAddress string) ConsulClient {
	// configure our consul client
	config := consul.Config{
		Address: consulAddress,
	}
	consulClient, err := consul.NewClient(&config)
	if err != nil {
		log.Fatal(err)
	}
	return ConsulClient{
		client: consulClient,
	}
}

// MonitorNode will monitor consul for changes in a node and, on changes, send back a list of service for that node
// that match our prefix
func (consulClient *ConsulClient) MonitorNode(node string, serviceOut chan<- services.NodeServices, cont <-chan bool) {
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
				log.Println(err)
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
						// grab our tags that have our prefix
						for _, tag := range service.Tags {
							if strings.HasPrefix(tag, viper.GetString("prefix")) {
								// parse our values
								values := strings.SplitN(strings.TrimPrefix(tag, viper.GetString("prefix")), ":", 2)
								foundServices.Services = append(foundServices.Services, services.Service{
									ConfigTemplate: values[0],
									DatadogType:    values[1],
									AgentService:   *service,
									Node:           node.Node.Node,
								})
							}
						}
					}
					// we always return if there was an updat since we need to know if services were removed
					serviceOut <- foundServices
				}
			}
		}
	}
}

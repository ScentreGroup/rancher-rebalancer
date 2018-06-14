package evencattle

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	r "github.com/ScentreGroup/rancher-rebalancer/rancher"
	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/v2"
)

//Function to return a list of services
func serviceIDList(client *rancher.RancherClient, projectID string) map[string]serviceDef {
	var service = map[string]serviceDef{}

	var rancherServices = map[string]string{
		"healthcheck":     "",
		"ipsec":           "",
		"network-manager": "",
		"scheduler":       "",
		"metadata":        "",
		"rancher-agent1":  ""}

	// get a list of hosts in this environment
	services, err := client.Service.List(nil)
	if err != nil {
		log.Error(err)
	}

	for _, h := range services.Data {
		if h.AccountId == projectID {
			if _, exists := rancherServices[h.Name]; exists {
				//Ignore it as its a Rancher Service
			} else {

				if len(h.InstanceIds) > 1 {
					var dat map[string]interface{}
					var affinity = false
					//Exclude the service is rebalance label is set to false
					labels, _ := json.Marshal(h.LaunchConfig.Labels)

					if err := json.Unmarshal(labels, &dat); err != nil {
						panic(err)
					}

					//Check for affinity rules
					for k := range dat {
						if strings.Index(k, "scheduler.affinity") > 0 {
							affinity = true
						}
					}

					if affinity {
						log.Debug("not balancing " + h.Name + " due to affinity rules")
					} else {
						var add = true

						// FIX!
						// I GUESS I"LL OPT IN?
						var opt = "IN"

						if opt == "IN" {
							if val, ok := dat["rebalance"]; ok {
								if val == "false" {
									add = false
								}
							} else {
								add = false
							}
						}

						if opt == "OUT" {
							if val, ok := dat["rebalance"]; ok {
								if val == "false" {
									add = false
								}
							}
						}

						if add {
							var tempService serviceDef
							tempService.id = h.Id
							tempService.name = h.Name
							tempService.instanceIds = h.InstanceIds
							service[h.Name+h.Id] = tempService
						}

					}
				} else {
					log.Debug("service " + h.Name + " only has 1 container, not balancing")
				}
			}
		}
	}
	return service
}

func serviceHosts(client *rancher.RancherClient, service serviceDef, hostList map[string]int, projectId string, mode string) int {
	hosts := make(map[string]int)
	containers := make(map[string]string)

	for h := range hostList {
		hosts[h] = 0
	}

	var instanceIDs = service.instanceIds

	for i := 0; i < len(instanceIDs); i++ {
		hostID := r.GetContainerHost(client, instanceIDs[i])
		for host := range hosts {
			if host == hostID {
				hosts[host] = hosts[host] + 1
			}
		}
		containers[instanceIDs[i]] = hostID
	}

	var balanced = evenLoad(client, projectId)

	//Find the variance
	var high = 0
	var low = 1000000

	//find out if the are evenly balanced
	for _, h := range hosts {
		if h > high {
			high = h
		}
		if h < low {
			low = h
		}
	}

	if high-low > 1 {

		var average = roundCount(len(hosts), len(instanceIDs))

		highest := ""
		for host := range hosts {
			if highest == "" {
				if hosts[host] > average {
					highest = host
				}
			}
			if hosts[host] > high { //always get the highest host
				highest = host
			}
		}

		if highest != "" {
			//Need to delete a container from this host
			//first find a container
			for instance := range containers {
				if containers[instance] == highest {
					if (mode == "aggressive") && (balanced) {
						//make the current host inactive
						currentHost, _ := client.Host.ById(highest)
						client.Host.ActionDeactivate(currentHost)
					}

					//Delete the container
					containerToDelete, err := client.Container.ById(instance)
					client.Container.Delete(containerToDelete)
					if err != nil {
						logrus.Error(err)
					}

					//Wait for 10 seconds to allow for allocations service to allocate new server
					time.Sleep(10 * time.Second)
					if (mode == "aggressive") && (balanced) {
						//make the current host inactive
						currentHost, _ := client.Host.ById(highest)
						client.Host.ActionActivate(currentHost)
					}
					break
				}
			}
		}
	} else {
		logrus.Info("Service is already balanced")
	}

	return 1
}

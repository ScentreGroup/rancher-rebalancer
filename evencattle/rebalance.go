package evencattle

import (
	r "github.com/ScentreGroup/rancher-rebalancer/rancher"
	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/v2"
)

// returns true or false if container service is balanced
func evenLoad(client *rancher.RancherClient, projectId string) bool {
	hostIds := map[string]int{}
	hostConMap := map[string]int{}

	// get a count of containers per host
	con, err := client.Container.List(nil)
	if err != nil {
		log.Error(err)
	}

	for _, d := range con.Data {
		hostConMap[d.HostId] = hostConMap[d.HostId] + 1
	}

	// get a list of active hosts in this environment
	hosts, err := client.Host.List(nil)
	for _, h := range hosts.Data {
		if h.AccountId == projectId && h.State == "active" {
			hostIds[h.Id] = hostConMap[h.Id]
		}
	}

	var high = 0
	var low = 1000000

	//find out if the are evenly balanced
	for _, h := range hostIds {
		if h > high {
			high = h
		}
		if h < low {
			low = h
		}
	}

	if high-low > 1 {
		return false //They aren't
	}

	return true //They are
}

// rebalances containers between nodes
func Rebalance(client *rancher.RancherClient, projectId string, mode string) {
	var balanced = false

	if mode != "aggressive" {
		balanced = evenLoad(client, projectId)
	}

	if !balanced {
		var hostList = r.GetHostIdsByProjectId(client, projectId)
		var serviceInstanceID = serviceIDList(client, projectId)

		for service := range serviceInstanceID {

			log.Info("processing service: " + service)

			serviceHosts(client, serviceInstanceID[service], hostList, projectId, mode)
		}
	} else {
		log.Info("server load balanced, aggressive mode would need to be used to enforce container balancing")
		return
	}
}

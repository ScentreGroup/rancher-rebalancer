package rancher

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/v2"
)

type Client struct {
	client *rancher.RancherClient
}

var withoutPagination *rancher.ListOpts

func CreateClient(url, accessKey, secretKey string) *rancher.RancherClient {
	client, err := rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Timeout:   time.Second * 5,
	})

	if err != nil {
		log.Errorf("failed to create a client for rancher, error: %s", err)
		os.Exit(1)
	}

	return client
}

// returns the project id of an environemnt
func GetProjectIdByName(client *rancher.RancherClient, name string) string {
	projects, err := client.Project.List(nil)
	if err != nil {
		log.Error("error getting project list")
	}

	for _, p := range projects.Data {
		if p.Name == name {
			return p.Id
		}
	}

	log.Error("project " + name + " not found")

	return "NotFound"
}

// returns a list of rancher host ids
func GetHostIdsByProjectId(client *rancher.RancherClient, projectID string) map[string]int {
	hostIds := map[string]int{}

	// get a list of hosts in the environment
	hosts, err := client.Host.List(nil)
	if err != nil {
		log.Error(err)
	}

	for _, h := range hosts.Data {
		if h.AccountId == projectID && h.State == "active" {
			hostIds[h.Id] = 0
		}
	}

	return hostIds
}

// returns the health state of an environemnt
func GetEnvironmentHealthByName(client *rancher.RancherClient, name string) string {
	projects, err := client.Project.List(nil)
	if err != nil {
		log.Error("error getting projects list")
	}

	for _, p := range projects.Data {
		if p.Name == name {
			return p.HealthState
		}
	}

	log.Error("environment " + name + " not found")

	return "NotFound"
}

// returns the host id of a container
func GetContainerHost(client *rancher.RancherClient, containerId string) string {
	services, err := client.Container.ById(containerId)
	if err != nil {
		log.Error(err)
	}

	return services.HostId
}

// returns the environment name of self via metadata service
func GetMetadataEnvironmentName(rancherMetadataUrl string) string {
	resp, err := http.Get(rancherMetadataUrl + "/latest/self/stack/environment_name")

	if err != nil {
		log.Errorf("rancher metadata not available ", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	return strings.Replace(string(body), "\"", "", -1)
}

// evacuates a rancher host
func EvacuateHostByName(client *rancher.RancherClient, hostName string) bool {
	// get a list of hosts
	hosts, err := client.Host.List(nil)
	if err != nil {
		log.Error(err)
	}

	for _, h := range hosts.Data {
		if h.Hostname == hostName {
			client.Host.ActionEvacuate(&h)
			return true
		}
	}

	return false
}

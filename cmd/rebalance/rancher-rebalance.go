package main

import (
	"os"
	"time"

	"github.com/ScentreGroup/rancher-rebalancer/evencattle"
	r "github.com/ScentreGroup/rancher-rebalancer/rancher"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	VERSION   = "v0.0.0-dev"
	projectId string
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "rancher-rebalance"
	app.Version = VERSION
	app.Usage = "re-balance cattle container services"
	app.Action = start
	app.Before = beforeApp
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "rancher-url",
			Value:  "http://localhost:8080/",
			Usage:  "full url of the rancher server",
			EnvVar: "CATTLE_URL",
		},
		cli.StringFlag{
			Name:   "rancher-access-key",
			Usage:  "rancher access Key",
			EnvVar: "CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "rancher-secret-key",
			Usage:  "rancher secret Key",
			EnvVar: "CATTLE_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "rancher-environment",
			Usage:  "rancher environment",
			EnvVar: "CATTLE_ENVIRONMENT",
		},
		cli.StringFlag{
			Name:   "rancher-metadata-url",
			Value:  "http://rancher-metadata",
			Usage:  "base url of the rancher metadata service",
			EnvVar: "RANCHER_METADATA_URL",
		},
		cli.StringFlag{
			Name:   "mode",
			Value:  "calm",
			Usage:  "the rebalancing strategy to use",
			EnvVar: "REBALANCE_MODE",
		},
		// not yet implemented
		cli.BoolFlag{
			Name:  "dry",
			Usage: "run in dry mode",
		},
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "run in debug mode",
		},
	}
	app.Run(os.Args)
}

func start(c *cli.Context) error {
	if len(c.String("rancher-url")) < 1 {
		log.Errorf("no rancher url provided")
		os.Exit(1)
	}

	if len(c.String("rancher-access-key")) < 1 {
		log.Errorf("no rancher access key provided")
		os.Exit(1)
	}

	if len(c.String("rancher-secret-key")) < 1 {
		log.Errorf("no rancher secret key provided")
		os.Exit(1)
	}

	log.Info("starting rancher rebalancer")

	rancherClient := r.CreateClient(c.String("rancher-url"),
		c.String("rancher-access-key"),
		c.String("rancher-secret-key"))

	if len(c.String("rancher-environment")) < 1 {
		// assume we should get the env that the container is running in
		environmentName := r.GetMetadataEnvironmentName(c.String("rancher-metadata-url"))
		log.Info("rancher environment set to " + environmentName)

		projectId = r.GetProjectIdByName(rancherClient, environmentName)
	}

	// start the health check server in a sub-process
	evencattle.StartHealthCheck()

	// main loop
	log.Debug("entering main loop")
	for {
		var returnCode = 0
		returnCode = evencattle.Rebalance(rancherClient, projectId, c.String("mode"))
		time.Sleep(time.Duration(returnCode) * time.Minute)
	}

	return nil
}

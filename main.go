package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/tcardonne/restic-controller/conf"
	"github.com/tcardonne/restic-controller/controller"
	"github.com/tcardonne/restic-controller/exporter"
)

func main() {
	configFile := flag.String("config", "config.yml", "Specify a configuration file to load")
	flag.Parse()

	config, err := conf.LoadConfiguration(*configFile)
	if err != nil {
		log.WithField("err", err).Fatal("Failed to load configuration")
	}
	if err := conf.ConfigureLogging(&config.Log); err != nil {
		log.WithField("err", err).Fatal("Failed to configure logging")
	}

	integrityController := controller.NewIntegrityController(config.Repositories)
	retentionController := controller.NewRetentionController(config.Repositories)
	exp := exporter.NewExporter(config.Exporter, config.Repositories, integrityController, retentionController)

	if err := integrityController.Start(); err != nil {
		log.WithField("err", err).Fatal("Failed to start integrity controller")
	}

	if err := retentionController.Start(); err != nil {
		log.WithField("err", err).Fatal("Failed to start retention controller")
	}

	if err := exp.ListenAndServe(); err != nil {
		log.WithField("err", err).Fatal("Failed starting http server")
	}
}

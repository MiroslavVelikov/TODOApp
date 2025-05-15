package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"project/graphql/graph/api"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	log.SetOutput(os.Stdout)
}

func main() {
	api.ServerHandler()
}

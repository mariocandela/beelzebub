package main

import (
	"os"

	"github.com/beelzebub-labs/beelzebub/v3/cli"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

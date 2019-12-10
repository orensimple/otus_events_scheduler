package main

import (
	"log"

	"github.com/orensimple/otus_events_scheduler/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

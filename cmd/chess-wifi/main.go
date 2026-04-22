package main

import (
	"log"

	"github.com/bssm-oss/chess-wifi/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}

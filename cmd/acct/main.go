package main

import (
	"log"
	"os"

	"auth.industrial-linguistics.com/accounting-ops/internal/cli"
)

func main() {
	app, err := cli.NewApp()
	if err != nil {
		log.Fatalf("initialise cli: %v", err)
	}
	code := app.Run(os.Args[1:])
	os.Exit(code)
}

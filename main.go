package main

import (
	"fmt"
	"os"

	"github.com/patrickdappollonio/cloudflare-cache-purger/cloudflare"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cloudflareToken, err := getAnyEnvironment("Cloudflare API Token", "INPUT_TOKEN", "TOKEN")
	if err != nil {
		return err
	}

	cloudflareZone, err := getAnyEnvironment("Cloudflare Zone ID", "INPUT_ZONE", "ZONE")
	if err != nil {
		return err
	}

	debug := os.Getenv("DEBUG") != ""

	cl := cloudflare.New(cloudflareToken)
	cl.SetDebug(debug)

	return cl.Clear(cloudflareZone)
}

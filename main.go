package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/webklex/gogeoip/src/app"
	"math/rand"
	"time"
)

//go:embed static
var staticFiles embed.FS

var buildNumber string
var buildVersion string

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := app.NewApplication(&app.Build{
		Number:  buildNumber,
		Version: buildVersion,
	}, flag.CommandLine, staticFiles).Start(); err != nil {
		fmt.Printf("[error] %s", err)
	}
}

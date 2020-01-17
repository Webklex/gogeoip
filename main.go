package main

import (
	"./server"
	"./utils/config"
	"flag"
	"fmt"
)

var buildNumber string
var buildVersion string

func main() {
	c := config.DefaultConfig()
	c.Load(c.File)

	c.AddFlags(flag.CommandLine)

	sv := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	c.Build = config.Build{
		Number: buildNumber,
		Version: buildVersion,
	}

	if *sv {
		fmt.Printf("geoIP version: %s\n", c.Build.Version)
		fmt.Printf("geoIP build number: %s\n", c.Build.Number)
		return
	}

	if c.SaveConfigFlag {
		if _, err := c.Save(); err != nil {
			print(err)
		}
	}

	s := server.NewServerConfig(c)
	s.Start()
}

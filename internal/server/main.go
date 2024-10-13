package main

import (
	"base-service/internal/conf"
	"base-service/internal/config"
	"base-service/internal/config/env"
	"base-service/internal/config/file"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "./../../config", "config path, eg: -conf config.yaml")
	godotenv.Load()
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			env.NewSource("BASE_"),
			file.NewSource(flagconf),
		),
	)

	if err := c.Load(); err != nil {
		panic(err)
	}

	var cf conf.Config
	if err := c.Scan(&cf); err != nil {
		panic(err)
	}

	fmt.Println(cf.GetServer())
}

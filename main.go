package main

import (
	"log"
	"strings"

	"github.com/icanwalkinyourmind/IntensiveGoPlateNumbers/confreader"
	"github.com/icanwalkinyourmind/IntensiveGoPlateNumbers/server"
)

func main() {
	var conf confreader.Config
	err := conf.ReadConfig("config.yaml", &conf)
	if err != nil {
		log.Println(err)
		conf = confreader.Config{Server: "", Port: "8000", Workers: 5}
	}
	log.Fatal(server.RunHTTPServer(strings.Join([]string{conf.Server, conf.Port}, ":"), conf.Workers))
}

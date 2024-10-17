package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/ry461ch/metric-collector/internal/app/agent"
	config "github.com/ry461ch/metric-collector/internal/config/agent"
)

func main() {
	agent := agent.New(config.New())
	go agent.Run()
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		fmt.Println(err)
	}
}

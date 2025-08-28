package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Rx-11/go-wasp/api"
	"github.com/Rx-11/go-wasp/config"
	"github.com/Rx-11/go-wasp/executor"
	"github.com/Rx-11/go-wasp/registry"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	var reg registry.Registry
	if cfg.RegistryType == "redis" {
		reg = registry.NewRedisRegistry()
	} else {
		reg = registry.NewMemoryRegistry()
	}

	disp := executor.NewDispatcher(
		reg,
		cfg.QueueSizePerFunction,
		cfg.DefaultFuncConcurrency,
		cfg.MaxNodeConcurrency,
		time.Duration(cfg.InvokeTimeoutMillis)*time.Millisecond,
	)

	srv := api.NewServer(reg, disp)
	app := srv.Routes()

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	fmt.Printf("go-wasp REST server running at %s\n", addr)
	log.Fatal(app.Listen(addr))
}

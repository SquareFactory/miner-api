package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	_ "embed"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/squarefactory/miner-api/api"
	"github.com/squarefactory/miner-api/autoswitch"
	"gopkg.in/yaml.v3"
)

//go:embed web/index.html
var f string

func main() {

	configPath := os.Getenv("CONFIG_PATH")
	cb, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	var config autoswitch.Config
	if err := yaml.Unmarshal(cb, &config); err != nil {
		log.Fatal(err)
	}

	switcher := &autoswitch.Switcher{
		Config: &config,
	}
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, r, f)
	})
	r.Post("/start", func(w http.ResponseWriter, r *http.Request) {

		api.MineStart(w, r, switcher)
	})
	r.Post("/stop", api.MineStop)
	r.Get("/health", api.Health)

	listenAddress := os.Getenv("LISTEN_ADDRESS")
	if len(listenAddress) == 0 {
		listenAddress = ":8080"
	}
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := http.Serve(l, r); err != nil {
			log.Fatal(err)
		}
	}()

	// context for the relaunch job goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(time.Duration(switcher.Config.General.PollingFrequency) * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Printf("autoswitch: restarting miners now")
			err := api.RestartMiners(ctx, switcher)
			if err != nil {
				log.Printf("failed to restart jobs: %s", err)
			}
		}
	}()

	wg.Wait()

}

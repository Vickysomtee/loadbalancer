package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/appleboy/loadbalancer-algorithms/weighted"
)

var (
	err        error
	mutex      sync.Mutex
	balance    weighted.RoundRobin
	configFile = flag.String("config", "config.json", "Location to config file")
)

type Config struct {
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []Server `json:"servers"`
}

type Server struct {
	Host           *url.URL `json:"host"`
	Url            string   `json:"url"`
	HealthCheckUrl string
	IsHealthy      bool
	Weight         int `json:"weight"`
}

func loadServers(servers []*Server) {
	for _, server := range servers {
		err := balance.AddServer(server.Host, server.Weight)

		if err != nil {
			fmt.Printf("Error adding %s to loadbalancer\n: %s", server.Url, err)
		} else {
			fmt.Printf("%s added to load balancer\n", server.Url)
		}
	}
}

func loadConfig(file string) (Config, error) {
	var config Config

	data, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func healthCheck(server *Server, healthCheckInterval time.Duration) {
	for range time.Tick(healthCheckInterval) {
		res, err := http.Head(server.HealthCheckUrl)
		mutex.Lock()
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Printf("%s is down\n", server.Url)
			server.IsHealthy = false
			balance.RemoveServer(server.Host)
		} else {
			server.IsHealthy = true
		}
		mutex.Unlock()
	}
}

func getNextServer(servers []*Server) *Server {
	url := balance.NextServer()
	if url == nil {
		return nil
	}

	for _, server := range servers {
		if url == server.Host {
			return server
		}
	}
	return nil
}

func main() {
	balance, err = weighted.New()
	if err != nil {
		panic(err)
	}
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Invalid health check interval: %s", err.Error())
	}

	var servers []*Server

	for _, srv := range config.Servers {
		u, _ := url.Parse(srv.Url)
		server := &Server{
			Host:           u,
			Url:            srv.Url,
			Weight:         srv.Weight,
			IsHealthy:      true,
			HealthCheckUrl: srv.HealthCheckUrl,
		}

		servers = append(servers, server)
		go healthCheck(server, healthCheckInterval)
	}

	loadServers(servers)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := getNextServer(servers)

		if server == nil {
			http.Error(w, "No available server", http.StatusServiceUnavailable)
			return
		}

		httputil.NewSingleHostReverseProxy(server.Host).ServeHTTP(w, r)
	})

	log.Println("Starting load balancer on port 7080")
	err = http.ListenAndServe(":7080", nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}

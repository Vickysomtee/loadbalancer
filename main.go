package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/appleboy/loadbalancer-algorithms/weighted"
)

// var balancer = balance.NewBalance()
var balance weighted.RoundRobin
var err error

type Config struct {
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []Server `json:"servers"`
}

type Server struct {
	Host   *url.URL `json:"host"`
	Url    string   `json:"url"`
	Weight int      `json:"weight"`
	// IsHealthy bool
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

func ReverseProxy(url *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(url)
}

func main() {
	balance, err = weighted.New()
	if err != nil {
		panic(err)
	}
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	var servers []*Server

	for _, srv := range config.Servers {
		u, _ := url.Parse(srv.Url)
		server := &Server{
			Host:   u,
			Url:    srv.Url,
			Weight: srv.Weight,
		}

		servers = append(servers, server)
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

	log.Println("Starting load balancer on port", config.Port)
	err = http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}

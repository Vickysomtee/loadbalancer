package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/mr-karan/balance"
)

var balancer = balance.NewBalance()

type Config struct {
	Port                string `json:"port"`
	HealthCheckInterval string `json:"healthCheckInterval"`
	Servers             []Url  `json:"servers"`
}

type Url struct {
	U string `json:"url"`
	// IsHealthy bool
	Weight int `json:"weight"`
}

type Server struct {
	URL *url.URL `json:"url"`
	// IsHealthy bool
}

func loadServers(config *Config) {
	for _, server := range config.Servers {
		err := balancer.Add(server.U, server.Weight)

		if err != nil {
			fmt.Printf("Error adding %s to loadbalancer\n: %s", server.U, err)
		} else {
			fmt.Printf("%s added to load balancer\n", server.U)
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

// round robin algorithm implementation to distribute load across servers
func getNextServer(servers []*Server) *Server {
	url := balancer.Get()
	for _, server := range servers {
		if url == server.URL.String() {
			return server
		}
	}
	return nil
}

func ReverseProxy(url *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(url)
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	loadServers(&config)

	var servers []*Server

	for _, u := range config.Servers {
		u, _ := url.Parse(u.U)
		server := &Server{
			URL: u,
		}

		servers = append(servers, server)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := getNextServer(servers)
		fmt.Println(servers)

		fmt.Println(server)

		// adding this header just for checking from which server the request is being handled.
		// this is not recommended from security perspective as we don't want to let the client know which server is handling the request.
		// w.Header().Add("X-Forwarded-Server", server.URL)
		httputil.NewSingleHostReverseProxy(server.URL).ServeHTTP(w, r)
	})

	log.Println("Starting load balancer on port", config.Port)
	err = http.ListenAndServe(config.Port, nil)
	if err != nil {
		log.Fatalf("Error starting load balancer: %s\n", err.Error())
	}
}

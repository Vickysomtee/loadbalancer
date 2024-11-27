A light weight loadbalancer for canary deployments. This utilizes the Weighted Round Robin algorithm to balance requests across resources based on defined weight.

Weights are defined with whole numbers which represents `nx10`, and the total weight of all servers added to the loadbalancer must sum up to 10.

Example
```
weight 5 represents 50% of traffic sent to the server
weight 3 represents 30% of traffic sent to the server
weight 2 represents 20% of traffic sent to the server

Total weights sums up to 10 which represent 100% of all traffic
```

# Getting Started

You can either run the loadbalancer with docker or clone the repo, build and run the binary with the docker method being the easiet. 

An important requirement is to have a `config.json` to load the servers and other details to the loadbalancer. Below is an example of the config file.

```json
{
  "healthCheckInterval": "2s",
  "servers": [
    {
      "url": "http://localhost:5100",
      "weight": 5,
      "healthCheckUrl": "http://localhost:5100/health"
    },
    {
      "url": "http://localhost:5200",
      "weight": 3,
      "healthCheckUrl": "http://localhost:5200/health"
    },
    {
      "url": "http://localhost:5300",
      "weight": 2,
      "healthCheckUrl": "http://localhost:5300/health"
    }
  ]
}
```

## Docker setup
Requirements
- Docker installed
- `config.json` file that will be mounted to the docker container

Run the command

```sh
docker run -p 7080:7080 -v ./config.json:/app/loadbalancer/config.json --name loadbalancer -d ghcr.io/vickysomtee/loadbalancer
 ```

## Clone repository
Requirements
- Go installed
- `config.json` file

Clone the Repository to your local directory
```sh
git clone https://github.com/Vickysomtee/loadbalancer.git
```

Build an executable
```sh
cd loadbalancer
go build
```
Run the executable 
```sh
./loadbalancer -config=config.json
```

Note: If you created your `config.json` file in the current working directory, there is no need to specify the config argument. Run the executable using

```sh
./loadbalancer
```

### Contributing
Contributions are welcome! Please submit a pull request or create an issue to contribute to this project.



 # Use an official Golang runtime as a parent image
 FROM golang:alpine

 # Set the working directory inside the container
 WORKDIR /app/loadbalancer

 # Copy the local package files to the container's workspace
 COPY . /app/loadbalancer

 # Build the Go application inside the container
 RUN go build

 EXPOSE 7080

 # Define the command to run your application
 ENTRYPOINT ["./loadbalancer -config=config.json"]
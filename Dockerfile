FROM golang:1.15.8

RUN apt-get update && apt-get install -y libpcap-dev

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY app .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080

VOLUME [ "/data" ]

# Run the executable
CMD ["k8s-arbitrage"]
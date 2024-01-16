# Using the official Golang 1.20 image as the base image
FROM golang:1.20

# the working directory inside the container
WORKDIR /app

# Copy the entire MEV Plus project to the container
COPY . .

RUN go build -o mevPlus mevPlus.go

EXPOSE 10000

CMD ["/bin/sh", "-c", "./mevPlus \
   -builderApi.listen-address $BUILDER_API_ADDRESS \
   -externalValidatorProxy.address $EXTERNAL_VALIDATOR_PROXY_ADDRESS \
   -k2.eth1-private-key $ETH1_PRIVATE_KEY \
   -k2.beacon-node-url $BEACON_NODE_URL \
   -k2.execution-node-url $EXECUTION_NODE_URL"]

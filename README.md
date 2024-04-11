# GRPC Proof of Concept (PoC)

This repository contains a Proof of Concept (PoC) for a gRPC service using Go. 

The purpose of this PoC is to demonstrate we can define a multi-tenant service and use a Header to identify the tenant 
without providing it in all requests.

In the case of the grpc-gateway, the idea is to evaluate if the tenant can be included in the path and then be injected
into the incoming headers of the gRPC service.


## Project Structure

The project is structured around the following key components:

- `proto/dummy/v1/service.proto`: This file defines the `DummyService` gRPC service, along with its request and response
  message types.

- `proto/buf.gen.yaml`: This file is used to generate Go code from the protobuf definitions using the `buf` tool.

- `proto/dummy/v1/service.yaml`: This file defines the HTTP/JSON to gRPC mapping for the `DummyService` using the
  gRPC-Gateway.

## How to Run

To run this PoC, you will need to have Go installed on your machine. Once you have Go installed, you can run the
following commands:

```bash
# Generate Go code from protobuf definitions
cd proto && make proto

# Run the server
go run cmd/main.go
```

This will start the gRPC server, and you can then use a gRPC client to interact with the `DummyService`.


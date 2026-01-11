# humaserverless

[![Go Reference](https://pkg.go.dev/badge/github.com/lucaspopp0/humaserverless.svg)](https://pkg.go.dev/github.com/lucaspopp0/humaserverless)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://go.dev/)

A Go library that seamlessly integrates [Huma v2](https://github.com/danielgtaylor/huma) with AWS Lambda, enabling you to build type-safe, OpenAPI-compliant HTTP APIs in serverless environments. This adapter handles the conversion between AWS API Gateway V2 HTTP events and standard HTTP requests/responses, allowing you to use Huma's powerful API framework in Lambda functions.

## Features

- 🔄 **Automatic Event Conversion** - Converts AWS Lambda API Gateway V2 events to standard HTTP requests and back
- 🎯 **Type-Safe Handlers** - Leverage Huma v2's type-safe request/response handling
- 📝 **OpenAPI Support** - Automatically generate OpenAPI specifications for your Lambda APIs
- 🔌 **Seamless Integration** - Works with existing Huma v2 APIs with minimal code changes
- 🛡️ **Error Handling** - Graceful error handling with proper HTTP status codes
- ⚡ **Zero Dependencies** - Lightweight adapter with minimal overhead

## Installation

```bash
go get github.com/lucaspopp0/humaserverless
```

## Quick Start

Here's a complete example of creating a Lambda handler with Huma:

```go
package main

import (
	"context"
	
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/lucaspopp0/humaserverless"
)

type GreetingRequest struct {
	Name string `path:"name" doc:"Name to greet"`
}

type GreetingResponse struct {
	Message string `json:"message"`
}

func main() {
	// Create a Chi router
	router := chi.NewMux()
	
	// Create a Huma API
	config := huma.DefaultConfig("My API", "1.0.0")
	api := humachi.New(router, config)
	
	// Register your operations
	huma.Register(api, huma.Operation{
		OperationID: "get-greeting",
		Method:      "GET",
		Path:        "/greet/{name}",
		Summary:     "Get a greeting",
	}, func(ctx context.Context, input *struct {
		Path GreetingRequest
	}) (*GreetingResponse, error) {
		return &GreetingResponse{
			Message: "Hello, " + input.Path.Name + "!",
		}, nil
	})
	
	// Create the Lambda handler
	handler := humaserverless.NewHttpHandler(api)
	
	// Start the Lambda function
	lambda.Start(handler)
}
```

### With Middleware

You can also use middleware to add cross-cutting concerns like logging, authentication, or request modification:

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/lucaspopp0/humaserverless"
)

func loggingMiddleware(next humaserverless.HttpHandlerFunc) humaserverless.HttpHandlerFunc {
	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		fmt.Printf("Request: %s %s\n", req.RequestContext.HTTP.Method, req.RawPath)
		resp, err := next(ctx, req)
		fmt.Printf("Response: %d\n", resp.StatusCode)
		return resp, err
	}
}

func main() {
	router := chi.NewMux()
	config := huma.DefaultConfig("My API", "1.0.0")
	api := humachi.New(router, config)
	
	// Register your operations...
	
	// Create handler with middleware
	baseHandler := humaserverless.NewHttpHandler(api)
	handler := loggingMiddleware(baseHandler)
	
	lambda.Start(handler)
}
```

### Deploying to AWS Lambda

1. Build your Lambda function:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
   ```

2. Create a deployment package with your `bootstrap` binary

3. Configure API Gateway V2 HTTP API to use your Lambda function

4. Your Huma API will automatically handle requests and generate OpenAPI documentation!

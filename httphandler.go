package humaserverless

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-lambda-go/events"
	"github.com/danielgtaylor/huma/v2"
)

type HttpHandlerFunc = func(
	ctx context.Context,
	request events.APIGatewayV2HTTPRequest,
) (events.APIGatewayV2HTTPResponse, error)

type HttpHandler[E, Req, Res any] struct {
	Env E

	Operation huma.Operation

	Init func(ctx context.Context) error

	Middlewares HttpMiddleware

	HandlerFunc func(
		ctx context.Context,
		request *Req,
	) (*Res, error)

	API huma.API
}

func (h *HttpHandler[E, Req, Res]) RegisterHandler(
	api huma.API,
) {
	huma.Register(
		api,
		h.Operation,
		h.HandlerFunc,
	)
}

func (h *HttpHandler[E, Req, Res]) HandleHttpEvent(
	ctx context.Context,
	event events.APIGatewayV2HTTPRequest,
) events.APIGatewayV2HTTPResponse {
	if h.Init == nil {
		h.Init = func(_ context.Context) error {
			return nil
		}
	}

	if h.Middlewares == nil {
		h.Middlewares = NoopHttpMiddleware
	}

	err := h.Init(ctx)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}
	}

	handleRequest := func(
		_ context.Context,
		event events.APIGatewayV2HTTPRequest,
	) (events.APIGatewayV2HTTPResponse, error) {
		// Convert the Lambda request to an HTTP request
		request, err := HttpEventToRequest(event)
		if err != nil {
			// If there is any issue, return a 500 error
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, nil
		}

		// Handle the request via Huma
		response := httptest.NewRecorder()
		h.API.Adapter().ServeHTTP(response, request)

		// Convert the HTTP response to a Lambda response
		return ResponseToHttpEvent(*response), nil
	}

	response, err := h.Middlewares(handleRequest)(ctx, event)
	if err != nil {
		// If there is any issue, return a 500 error
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}
	}

	return response
}

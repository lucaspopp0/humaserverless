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

func NewHttpHandler(api huma.API) HttpHandlerFunc {
	return func(
		ctx context.Context,
		requestEvent events.APIGatewayV2HTTPRequest,
	) (events.APIGatewayV2HTTPResponse, error) {
		// Convert the Lambda request to an HTTP request
		request, err := HttpEventToRequest(requestEvent)
		if err != nil {
			// If there is any issue, return a 500 error
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, nil
		}

		// Handle the request via Huma
		response := httptest.NewRecorder()
		api.Adapter().ServeHTTP(response, request)

		// Convert the HTTP response to a Lambda response
		return ResponseToHttpEvent(*response), nil
	}
}

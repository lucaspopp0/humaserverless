package humaserverless

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Convert the API Gateway HTTP request event into a HTTP request
func HttpEventToRequest(
	event events.APIGatewayV2HTTPRequest,
) (*http.Request, error) {
	routePieces := strings.Split(event.RouteKey, " ")
	httpMethod := routePieces[0]

	body := strings.NewReader(event.Body)

	requestURL := event.RawPath
	if event.RequestContext.Stage != "" {
		requestURL = strings.TrimPrefix(requestURL,
			fmt.Sprintf("/%s", event.RequestContext.Stage))
	}

	if event.RawQueryString != "" {
		requestURL = fmt.Sprintf("%s?%s", requestURL, event.RawQueryString)
	}

	request, err := http.NewRequest(httpMethod, requestURL, body)
	if err != nil {
		return nil, err
	}

	for headerName, headerVal := range event.Headers {
		request.Header.Set(headerName, headerVal)
	}

	return request, nil
}

// Convert an HTTP response to an API Gateway HTTP response event
func ResponseToHttpEvent(
	response httptest.ResponseRecorder,
) events.APIGatewayV2HTTPResponse {
	event := events.APIGatewayV2HTTPResponse{
		StatusCode: response.Code,
		Body:       response.Body.String(),

		Headers:           map[string]string{},
		MultiValueHeaders: map[string][]string{},
	}

	for headerName, headerValues := range response.Header() {
		if len(headerValues) > 0 {
			event.Headers[headerName] = headerValues[len(headerValues)-1]
			event.MultiValueHeaders[headerName] = headerValues
		}
	}

	return event
}

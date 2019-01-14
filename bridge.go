package lambda_http_bridge

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"net/http"
	"net/url"
)

type LambdaProxyHttpBridge struct {
	handler http.Handler
}

func ServeLambda(handler http.Handler) {
	lambda.StartHandler(LambdaProxyHttpBridge{handler})
}

func (b LambdaProxyHttpBridge) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	var (
		responseBytes []byte
		request       *http.Request
	)

	parsed := events.APIGatewayProxyRequest{}
	err := json.Unmarshal(payload, &parsed)

	if err == nil {
		if request, err = buildBridgeRequest(ctx, &parsed); err == nil {
			captor := captureResponseWriter{}
			b.handler.ServeHTTP(&captor, request)
			response := captor.toLambdaResponse()

			responseBytes, err = json.Marshal(&response)
		}

	}

	return responseBytes, err
}

func buildBridgeRequest(ctx context.Context, request *events.APIGatewayProxyRequest) (rq *http.Request, err error) {
	var (
		uri  *url.URL
		body []byte
	)
	if uri, err = url.Parse(request.Path); err == nil {
		header := http.Header{}

		for k, v := range request.Headers {
			header.Add(k, v)
		}

		if request.IsBase64Encoded {
			body, err = base64.StdEncoding.DecodeString(request.Body)
		} else {
			body = []byte(request.Body)
		}

		rq = (&http.Request{
			Method:        request.HTTPMethod,
			URL:           uri,
			RequestURI:    request.Path,
			Header:        header,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
			ContentLength: int64(len(body)),
		}).WithContext(ctx)
	}

	return

}

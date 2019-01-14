package lambda_http_bridge

import (
	"bytes"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"net/http"
	"strconv"
)

type captureResponseWriter struct {
	headers       http.Header
	statusCode    int
	content       bytes.Buffer
	contentLength uint64
}

func (c *captureResponseWriter) Header() http.Header {
	if c.headers == nil {
		c.headers = make(map[string][]string)
	}

	return c.headers
}

func (c *captureResponseWriter) Write(in []byte) (int, error) {
	if c.statusCode == 0 {
		c.statusCode = 200
	}

	c.contentLength += uint64(len(in))
	got, err := c.content.Write(in)

	return got, err
}

func (c *captureResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
}

func needsMinimalEscaping(in []byte) bool {
	for _, next := range in {
		if next < 32 || next >= 126 || next == '\r' || next == '\n' || next == '\t' {
			return false
		}
	}

	return true
}

func (c *captureResponseWriter) toLambdaResponse() events.APIGatewayProxyResponse {
	singleHeaders := map[string]string{
		"Content-Length": strconv.FormatUint(c.contentLength, 10),
	}

	for k, v := range c.headers {
		singleHeaders[k] = v[0]
		if len(v) > 1 {
			log.Println("Multi-value header encountered:", k, "dropping all but first occurrance", v[1:])
		}
	}

	contentBytes := c.content.Bytes()
	needsBase64 := !needsMinimalEscaping(contentBytes)
	body := ""

	if needsBase64 {
		body = base64.StdEncoding.EncodeToString(contentBytes)
	} else {
		body = string(contentBytes)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      c.statusCode,
		Headers:         singleHeaders,
		IsBase64Encoded: needsBase64,
		Body:            body,
	}

}

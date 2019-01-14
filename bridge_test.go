package lambda_http_bridge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	plainProxyBody = `{
  "body": "Mali Mirassin",
  "resource": "/{proxy+}",
  "path": "/path/to/resource",
  "httpMethod": "POST",
  "isBase64Encoded": false,
  "queryStringParameters": {
    "foo": "bar"
  },
  "pathParameters": {
    "proxy": "/path/to/resource"
  },
  "stageVariables": {
    "baz": "qux"
  },
  "headers": {
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
    "Accept-Encoding": "gzip, deflate, sdch",
    "Accept-Language": "en-US,en;q=0.8",
    "Cache-Control": "max-age=0",
    "Content-Type": "appliction/json",
    "CloudFront-Forwarded-Proto": "https",
    "CloudFront-Is-Desktop-Viewer": "true",
    "CloudFront-Is-Mobile-Viewer": "false",
    "CloudFront-Is-SmartTV-Viewer": "false",
    "CloudFront-Is-Tablet-Viewer": "false",
    "CloudFront-Viewer-Country": "US",
    "Host": "1234567890.execute-api.eu-west-1.amazonaws.com",
    "Upgrade-Insecure-Requests": "1",
    "User-Agent": "Custom User Agent String",
    "Via": "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)",
    "X-Amz-Cf-Id": "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==",
    "X-Forwarded-For": "127.0.0.1, 127.0.0.2",
    "X-Forwarded-Port": "443",
    "X-Forwarded-Proto": "https"
  },
  "requestContext": {
    "accountId": "123456789012",
    "resourceId": "123456",
    "stage": "prod",
    "requestId": "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
    "requestTime": "09/Apr/2015:12:34:56 +0000",
    "requestTimeEpoch": 1428582896000,
    "identity": {
      "cognitoIdentityPoolId": null,
      "accountId": null,
      "cognitoIdentityId": null,
      "caller": null,
      "accessKey": null,
      "sourceIp": "127.0.0.1",
      "cognitoAuthenticationType": null,
      "cognitoAuthenticationProvider": null,
      "userArn": null,
      "userAgent": "Custom User Agent String",
      "user": null
    },
    "path": "/prod/path/to/resource",
    "resourcePath": "/{proxy+}",
    "httpMethod": "POST",
    "apiId": "1234567890",
    "protocol": "HTTP/1.1"
  }
}`
	proxyExampleBody = `{
  "body": "eyJ0ZXN0IjoiYm9keSJ9",
  "resource": "/{proxy+}",
  "path": "/path/to/resource",
  "httpMethod": "POST",
  "isBase64Encoded": true,
  "queryStringParameters": {
    "foo": "bar"
  },
  "pathParameters": {
    "proxy": "/path/to/resource"
  },
  "stageVariables": {
    "baz": "qux",
	"My-Stage-Variable": "yellowsubmarine"
  },
  "headers": {
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
    "Accept-Encoding": "gzip, deflate, sdch",
    "Accept-Language": "en-US,en;q=0.8",
    "Cache-Control": "max-age=0",
    "Content-Type": "appliction/json",
    "CloudFront-Forwarded-Proto": "https",
    "CloudFront-Is-Desktop-Viewer": "true",
    "CloudFront-Is-Mobile-Viewer": "false",
    "CloudFront-Is-SmartTV-Viewer": "false",
    "CloudFront-Is-Tablet-Viewer": "false",
    "CloudFront-Viewer-Country": "US",
    "Host": "1234567890.execute-api.eu-west-1.amazonaws.com",
    "Upgrade-Insecure-Requests": "1",
    "User-Agent": "Custom User Agent String",
    "Via": "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)",
    "X-Amz-Cf-Id": "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==",
    "X-Forwarded-For": "127.0.0.1, 127.0.0.2",
    "X-Forwarded-Port": "443",
    "X-Forwarded-Proto": "https"
  },
  "requestContext": {
    "accountId": "123456789012",
    "resourceId": "123456",
    "stage": "prod",
    "requestId": "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
    "requestTime": "09/Apr/2015:12:34:56 +0000",
    "requestTimeEpoch": 1428582896000,
    "identity": {
      "cognitoIdentityPoolId": null,
      "accountId": null,
      "cognitoIdentityId": null,
      "caller": null,
      "accessKey": null,
      "sourceIp": "127.0.0.1",
      "cognitoAuthenticationType": null,
      "cognitoAuthenticationProvider": null,
      "userArn": null,
      "userAgent": "Custom User Agent String",
      "user": null
    },
    "path": "/prod/path/to/resource",
    "resourcePath": "/{proxy+}",
    "httpMethod": "POST",
    "apiId": "1234567890",
    "protocol": "HTTP/1.1"
  }
}`
	single = "X-Single-Header"
	multi  = "X-Multi-Header"
)

func verifyHttpHeadersAndPathAsExpected(rw http.ResponseWriter, rq *http.Request) {
	head := rq.Header
	if head.Get("Accept") == "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8" &&
		head.Get("Accept-Encoding") == "gzip, deflate, sdch" &&
		head.Get("Accept-Language") == "en-US,en;q=0.8" &&
		head.Get("Cache-Control") == "max-age=0" &&
		head.Get("Content-Type") == "appliction/json" &&
		head.Get("CloudFront-Forwarded-Proto") == "https" &&
		head.Get("CloudFront-Is-Desktop-Viewer") == "true" &&
		head.Get("CloudFront-Is-Mobile-Viewer") == "false" &&
		head.Get("CloudFront-Is-SmartTV-Viewer") == "false" &&
		head.Get("CloudFront-Is-Tablet-Viewer") == "false" &&
		head.Get("CloudFront-Viewer-Country") == "US" &&
		head.Get("Host") == "1234567890.execute-api.eu-west-1.amazonaws.com" &&
		head.Get("Upgrade-Insecure-Requests") == "1" &&
		head.Get("User-Agent") == "Custom User Agent String" &&
		head.Get("Via") == "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)" &&
		head.Get("X-Amz-Cf-Id") == "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==" &&
		head.Get("X-Forwarded-For") == "127.0.0.1, 127.0.0.2" &&
		head.Get("X-Forwarded-Port") == "443" &&
		head.Get("X-Forwarded-Proto") == "https" &&
		head.Get("X-Api-Gateway-Stage-Variable-Baz") == "qux" &&
		head.Get("X-Api-Gateway-Stage-Variable-My-Stage-Variable") == "yellowsubmarine" &&
		rq.URL.Path == "/path/to/resource" {
		rw.WriteHeader(204)
	} else {
		rw.WriteHeader(500)
	}
}

func noContentHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add(multi, "abc")
	rw.Header().Add(multi, "def")
	rw.Header().Add(single, "Hello World")
	rw.WriteHeader(204)
}

func echoWriter(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Add("Content-Type", "text/plain")
	rest, _ := ioutil.ReadAll(rq.Body)
	_, _ = rw.Write([]byte("Hello "))
	_, _ = rw.Write(rest)
}

func binaryWriter(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/octet-stream")
	_, _ = rw.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
}

func TestSimpleBridgeInvokePlain(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(echoWriter)}
	resultBytes, err := sut.Invoke(context.Background(), []byte(plainProxyBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	if result.StatusCode != 200 || result.IsBase64Encoded || result.Body != "Hello Mali Mirassin" {
		t.Error("Mismatched response:", result)
	}
}

func TestSimpleBridgeInvoke(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(noContentHandler)}
	resultBytes, err := sut.Invoke(context.Background(), []byte(proxyExampleBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	if result.StatusCode != 204 || result.Body != "" || result.IsBase64Encoded {
		t.Error("Mismatched response:", result)
	}
}

func TestHeaderSimplification(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(noContentHandler)}
	resultBytes, err := sut.Invoke(context.Background(), []byte(proxyExampleBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	singleV := result.Headers[single]
	multiV := result.Headers[multi]

	if singleV != "Hello World" || !(multiV == "abc" || multiV == "def") {
		t.Error("Headers mismatched", singleV, multiV)
	}
}

func TestBodyEncoding(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(echoWriter)}

	resultBytes, err := sut.Invoke(context.Background(), []byte(proxyExampleBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	if string(result.Body) != `Hello {"test":"body"}` {
		t.Error("Decoded body was ", result.Body)
	}
}

func TestBase64Body(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(binaryWriter)}

	resultBytes, err := sut.Invoke(context.Background(), []byte(proxyExampleBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	if !result.IsBase64Encoded {
		t.Error("binary result was not base64 encoded")
	}

	var decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(result.Body)
	if err != nil {
		t.Error("base64 decode failed", err)
	}

	for i := range decoded {
		if decoded[i] != byte(i) {
			t.Error("Decode error at", i)
		}
	}
}

func TestHeaderForwarding(t *testing.T) {
	sut := LambdaProxyHttpBridge{http.HandlerFunc(verifyHttpHeadersAndPathAsExpected)}
	resultBytes, err := sut.Invoke(context.Background(), []byte(proxyExampleBody))

	if err != nil {
		t.Error(err, "was not nil")
	}

	var result events.APIGatewayProxyResponse
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		t.Error(err, "reported during result unmarshal")
	}

	if result.StatusCode != 204 {
		t.Error("Status code was", result.StatusCode, "instead of 204")
	}
}

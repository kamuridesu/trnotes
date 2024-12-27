package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	Method         string
	Url            string
	Headers        map[string]string
	Body           string
	ExpectedStatus int
}

func New(method, url string, expectedStatus int) *Request {
	return &Request{
		Method:         method,
		Url:            url,
		Body:           "",
		ExpectedStatus: expectedStatus,
	}
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	r.Headers = headers
	return r
}

func (r *Request) SetBody(body string) *Request {
	r.Body = body
	return r
}

func (r *Request) Send() (string, error) {
	var rawBody io.Reader
	if r.Body != "" {
		rawBody = strings.NewReader(r.Body)
	}
	req, err := http.NewRequest(r.Method, r.Url, rawBody)
	if err != nil {
		return "", err
	}
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != r.ExpectedStatus {
		return "", fmt.Errorf("unexpected status code, expected %d, got %d, body is %s", r.ExpectedStatus, res.StatusCode, body)
	}
	return string(body), nil
}

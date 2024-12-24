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
	req            *http.Request
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

func (r *Request) prepare() error {
	var body io.Reader
	if r.Body != "" {
		body = strings.NewReader(r.Body)
	}
	req, err := http.NewRequest(r.Method, r.Url, body)
	if err != nil {
		return err
	}
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}
	r.req = req
	return nil
}

func (r *Request) Send() (string, error) {
	err := r.prepare()
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(r.req)
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

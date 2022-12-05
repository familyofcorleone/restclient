package restclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	httpClientTimeout = 15 * time.Second
)

var (
	ErrUnexpectedResponse = errors.New("the API is currently unavailable")
)

type Client interface {
	Request(v interface{}, method, path string, data interface{}) error
}

type DefaultClient struct {
	HTTPClient *http.Client
	DebugLog   *log.Logger
}

type contentType string

const (
	contentTypeEmpty          contentType = ""
	contentTypeJSON           contentType = "application/json"
	contentTypeFormURLEncoded contentType = "application/x-www-form-urlencoded"
)

type errorReader func([]byte) error

var customErrorReader errorReader

func New() *DefaultClient {
	return &DefaultClient{
		HTTPClient: &http.Client{
			Timeout: httpClientTimeout,
		},
	}
}

func (c *DefaultClient) Request(v interface{}, method, path string, data interface{}, m map[string]string) error {
	uri, err := url.Parse(path)
	if err != nil {
		return err
	}
	//todo sonra unutma sil proxyi
	// proxyUrl, err := url.Parse("http://127.0.0.1:8888")
	// http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

	body, contentType, err := prepareRequestBody(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(method, uri.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	for k, v := range m {
		request.Header.Set(k, v)
	}

	if contentType != contentTypeEmpty {
		request.Header.Set("Content-Type", string(contentType))
	}

	if c.DebugLog != nil {
		if data != nil {
			c.DebugLog.Printf("HTTP REQUEST: %s %s %s", method, uri.String(), body)
		} else {
			c.DebugLog.Printf("HTTP REQUEST: %s %s", method, uri.String())
		}
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if c.DebugLog != nil {
		c.DebugLog.Printf("HTTP RESPONSE: %s", string(responseBody))
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		if v != nil {
			if err := json.Unmarshal(responseBody, &v); err != nil {
				return fmt.Errorf("could not decode response JSON, %s: %v", string(responseBody), err)
			}
		}
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusInternalServerError:
		return ErrUnexpectedResponse
	default:
		if customErrorReader != nil {
			return customErrorReader(responseBody)
		}

		return defaultErrorReader(responseBody)
	}
}

func defaultErrorReader(b []byte) error {
	var errorResponse ErrorResponse

	if err := json.Unmarshal(b, &errorResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response json %s, error: %v", string(b), err)
	}

	return errorResponse
}

func prepareRequestBody(data interface{}) ([]byte, contentType, error) {
	switch data := data.(type) {
	case nil:
		return nil, contentTypeEmpty, nil
	case string:
		return []byte(data), contentTypeFormURLEncoded, nil
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return nil, "", err
		}
		return b, contentTypeJSON, nil
	}
}

package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"time"
)

var TIMEOUT = 30

type Auth struct {
	Username string
	Password string
}

func HTTPGet(uri string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return HTTPRequestWithoutBody(uri, params, headers, http.MethodGet)
}

func HTTPSGet(uri string, params map[string]string, headers map[string]string, caPath string) (*http.Response, error) {
	return HTTPSRequestWithoutBody(uri, params, headers, http.MethodGet, caPath)
}

func HTTPDelete(uri string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return HTTPRequestWithoutBody(uri, params, headers, http.MethodDelete)
}

func HTTPSDelete(uri string, params map[string]string, headers map[string]string, caPath string) (*http.Response, error) {
	return HTTPSRequestWithoutBody(uri, params, headers, http.MethodDelete, caPath)
}

func HTTPPost(url string, body interface{}, params map[string]string, headers map[string]string) (*http.Response, error) {
	return HTTPRequestWithBody(url, body, params, headers, http.MethodPost)
}

func HTTPSPost(url string, body interface{}, params map[string]string, headers map[string]string, caPath string) (*http.Response, error) {
	return HTTPSRequestWithBody(url, body, params, headers, http.MethodPost, caPath)
}

func HTTPPatch(url string, body interface{}, params map[string]string, headers map[string]string) (*http.Response, error) {
	return HTTPRequestWithBody(url, body, params, headers, http.MethodPatch)
}

func HTTPSPatch(url string, body interface{}, params map[string]string, headers map[string]string, caPath string) (*http.Response, error) {
	return HTTPSRequestWithBody(url, body, params, headers, http.MethodPatch, caPath)
}

func HTTPPut(url string, body interface{}, params map[string]string, headers map[string]string) (*http.Response, error) {
	return HTTPRequestWithBody(url, body, params, headers, http.MethodPost)
}

func HTTPRequestWithoutBody(url string, params map[string]string, headers map[string]string, method string) (*http.Response, error) {
	client := NewClient(time.Duration(TIMEOUT) * time.Second)
	return requestWithoutBody(url, params, headers, method, client)
}

func HTTPSRequestWithoutBody(url string, params map[string]string, headers map[string]string, method string, caPath string) (*http.Response, error) {
	client := NewHttpsClient(time.Duration(TIMEOUT)*time.Second, caPath)
	return requestWithoutBody(url, params, headers, method, client)
}

func requestWithoutBody(uri string, params map[string]string, headers map[string]string, method string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest(method, uri, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	q := req.URL.Query()

	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}

	return client.Do(req)
}

func HTTPRequestWithBody(url string, body interface{}, params map[string]string, headers map[string]string, method string) (*http.Response, error) {
	client := NewClient(time.Duration(TIMEOUT) * time.Second)
	return requestWithBody(url, body, params, headers, method, client)
}

func HTTPSRequestWithBody(url string, body interface{}, params map[string]string, headers map[string]string, method string, caPath string) (*http.Response, error) {
	client := NewHttpsClient(time.Duration(TIMEOUT)*time.Second, caPath)
	return requestWithBody(url, body, params, headers, method, client)
}

func requestWithBody(url string, body interface{}, params map[string]string, headers map[string]string, method string, client *http.Client) (*http.Response, error) {
	var bodyJSON []byte
	var req *http.Request
	if body != nil {
		var err error
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			log.Println(err)
			return nil, errors.New("http post body to json failed")
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyJSON))

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf(fmt.Sprintf("new request is fail: %v \n", err))
	}

	if method == http.MethodPatch {
		req.Header.Set("Content-type", "application/json-patch+json")
	} else {
		req.Header.Set("Content-type", "application/json;charset=utf-8")
	}

	q := req.URL.Query()

	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	return client.Do(req)
}

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, timeout)
				if err != nil {
					return nil, err
				}
				err = conn.SetDeadline(time.Now().Add(timeout))
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			ResponseHeaderTimeout: timeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: timeout * 2,
	}
}

func NewHttpsClient(timeout time.Duration, caPath string) *http.Client {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, timeout)
				if err != nil {
					return nil, err
				}
				err = conn.SetDeadline(time.Now().Add(timeout))
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			ResponseHeaderTimeout: timeout,
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: timeout * 2,
	}
}

func newRequest(url string, body interface{}, params map[string]string, headers map[string]string, method string) (*http.Request, error) {
	var read io.Reader
	var req *http.Request
	if body != nil {
		switch body.(type) {
		case multipart.File:
			file, err := ioutil.ReadAll(body.(multipart.File))
			if err != nil {
				return nil, err
			}
			read = bytes.NewReader(file)
		case []byte:
			read = bytes.NewReader(body.([]byte))
		default:
			bodyJSON, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			read = bytes.NewBuffer(bodyJSON)
		}
	}

	req, err := http.NewRequest(method, url, read)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail: %v \n")
	}

	q := req.URL.Query()

	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}

	return req, nil
}

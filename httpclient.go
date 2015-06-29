package sapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	POST   = "POST"
	GET    = "GET"
	UPDATE = "PUT"
	DELETE = "DELETE"
)

type HttpAgent struct {
	Url       string
	Method    string
	Data      interface{}
	DataType  string
	Client    *http.Client
	Transport *http.Transport
	Header    map[string]string
}

func NewHttpAgent() *HttpAgent {
	return &HttpAgent{
		Header:    make(map[string]string),
		Client:    &http.Client{},
		Transport: &http.Transport{},
		DataType:  "json",
	}
}

func (h *HttpAgent) Clear() *HttpAgent {
	h.Url = ""
	h.Method = ""
	h.DataType = "json"
	h.Header = make(map[string]string)
	return h
}

func (h *HttpAgent) Get(url string) *HttpAgent {
	h.Clear()
	h.Url = url
	h.Method = GET
	return h
}

func (h *HttpAgent) Post(url string) *HttpAgent {
	h.Clear()
	h.Url = url
	h.Method = POST
	return h
}

func (h *HttpAgent) Put(url string) *HttpAgent {
	h.Clear()
	h.Url = url
	h.Method = UPDATE
	return h
}

func (h *HttpAgent) Delete(url string) *HttpAgent {
	h.Clear()
	h.Url = url
	h.Method = DELETE
	return h
}

func (h *HttpAgent) SetHeader(para string, value string) *HttpAgent {
	h.Header[para] = value
	return h
}

func (h *HttpAgent) Timeout(timeout time.Duration) *HttpAgent {
	h.Transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, nil
	}
	return h
}

func (h *HttpAgent) ReqData(data interface{}) *HttpAgent {
	h.Data = data
	return h
}

func (h *HttpAgent) Issue() (*http.Response, string, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	reqBody, err := json.Marshal(h.Data)
	if err != nil {
		return nil, "", err
	}
	req, err = http.NewRequest(h.Method, h.Url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", err
	}

	h.SetHeader("Content-Type", "application/json")
	for para, value := range h.Header {
		req.Header.Set(para, value)
	}
	h.Client.Transport = h.Transport

	resp, err = h.Client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	bodyString := string(body)

	return resp, bodyString, nil
}

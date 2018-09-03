package core

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	// RequestGET defines a GET HTTP request method.
	RequestGET = "GET"
)

// Request implements a generic HTTP request handler. It will invoke a request
// of type method to the given url with an optional payload. The raw response
// body and any error will be returned.
func Request(url, method string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	rawBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	res.Body.Close()
	return rawBody, nil
}

// ClientManager implements a simple round-robin load balancing client manager.
type ClientManager struct {
	mu      sync.Mutex
	index   int
	clients []string
}

// NewClientManager returns a reference to a new initialized ClientManager with
// a given list of clients.
func NewClientManager(clients []string) *ClientManager {
	return &ClientManager{clients: clients}
}

// Next returns the next client to be used from the client manager. Each client
// is round-robin load balanced.
func (cm *ClientManager) Next() string {
	cm.mu.Lock()
	client := cm.clients[cm.index]
	cm.index = (cm.index + 1) % len(cm.clients)
	cm.mu.Unlock()

	return client
}

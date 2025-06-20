package quester

import (
	"net/http"
)

type Hooks interface {
	PreRequest(req *http.Request) error
	PostResponse(res *http.Response) error
}

type DefaultHooks struct{}

func (d *DefaultHooks) PreRequest(req *http.Request) error {
	return nil
}

func (d *DefaultHooks) PostResponse(res *http.Response) error {
	return nil
}

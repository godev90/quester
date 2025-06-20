package http

import (
	"net/http"
)

type Response struct {
	Status     int
	StatusText string
	Headers    http.Header
	Body       any
}

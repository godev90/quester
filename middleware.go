package http

import (
	"log"
	"net/http"
)

func LogRequest(req *http.Request) {
	log.Printf("[Request] %s %s", req.Method, req.URL.String())
	for k, v := range req.Header {
		log.Printf("Header: %s = %v", k, v)
	}
}

func LogResponse(res *http.Response) {
	log.Printf("[Response] %d %s", res.StatusCode, res.Status)
	for k, v := range res.Header {
		log.Printf("Header: %s = %v", k, v)
	}
}

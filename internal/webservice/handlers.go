package webservice

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

// rootHandler will return a list of available endpoints
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	bw := bufio.NewWriter(w)
	bw.WriteString("\n\nAvailable Service Endpoints\n===========================\n\n")
	for _, ep := range s.httpEndpoints {
		fmt.Fprintf(bw, "curl -X %s localhost:%d%s\n", ep.Method, s.port, ep.Path)
	}
	bw.Flush()
}

func (s *Server) codeHandler(code int) http.HandlerFunc {
	if code < 100 || code > 599 {
		// panic before invalid call to WriteHeader (See https://golang.org/src/net/http/server.go#L1078)
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := delayResponse(r)
		if err != nil {
			defer fmt.Fprintf(w, "invalid delay: %s\n", err.Error())
		}
		w.WriteHeader(code)
		fmt.Fprintf(w, "%d\n", code)
		// fmt.Printf("%d\n", code)
	}
}

func delayResponse(r *http.Request) error {
	delay := r.URL.Query().Get("delay")
	if len(delay) == 0 {
		return nil
	}
	d, err := time.ParseDuration(delay)
	if err != nil {
		return err
	}
	time.Sleep(d)
	return nil
}

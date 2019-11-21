package main

import "github.com/jimmyjames85/metrics/internal/webservice"

func main() {
	port := 8080
	s := webservice.New(port)
	err := s.Serve()
	if err != nil {
		panic(err)
	}
}

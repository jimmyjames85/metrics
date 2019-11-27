package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jimmyjames85/metrics/internal/webservice"
)

func main() {
	port := 5555

	p := os.Getenv("WEBSERVICE_PORT")
	if p != "" {
		var err error
		port, err = strconv.Atoi(p)
		if err != nil || port <= 0 {
			fmt.Fprintf(os.Stderr, "invalid port specification for WEBSERVICE_PORT: %s\n", err.Error())
			os.Exit(-1)
		}
	}

	s := webservice.New(port)
	err := s.Serve()
	if err != nil {
		panic(err)
	}
}

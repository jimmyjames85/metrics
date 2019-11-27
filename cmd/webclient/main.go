package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func toCURL(req *http.Request) (string, error) {
	bw := &bytes.Buffer{}
	fmt.Fprintf(bw, "curl -X %s", req.Method)
	for h, vals := range req.Header {
		for _, v := range vals {
			fmt.Fprintf(bw, " -H '%s: %s'", h, v)
		}
	}
	fmt.Fprintf(bw, " %s", req.URL.String())
	if req.Body != nil {
		defer req.Body.Close()
		byts, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}
		req.Body = ioutil.NopCloser(strings.NewReader(string(byts)))
		fmt.Fprintf(bw, " -d %q", string(byts))
	}
	return bw.String(), nil
}

func submitRequest(cli *http.Client, req *http.Request, printCurl bool) ([]byte, error) {
	if printCurl {
		curl, err := toCURL(req)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s\n", curl)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		err = fmt.Errorf("non-200 return code: %d, response: %s", resp.StatusCode, string(byt))
	}

	return byt, err
}

var opts struct {
	verbose     bool
	concurrency int           // concurent number of users making requests
	period      time.Duration // how long to run tests for
	url         *url.URL
	method      string // GET, POST, PUT, etc
}

func usage(code int) {
	fmt.Printf(`Makes and executes an http request repeatedly against a URL.

Usage:
  %s [flags] URL

Flags:

        -t, --time <duration>   How long to repeat the test for. Valid
				time units are "ns", "us" (or "µs"),
				"ms", "s", "m", "h".

        -c, --concurrency	How many concurrent requests to make (default 1)
        -X, --request		Specifies the request method to use (default GET)

        -h, --help              Displays this message

        -v, --verbose		Displays the http requests being in curl
				format. If concurrency is greater than
				one, only displays the curl for the
				first worker.
`, os.Args[0])

	os.Exit(code)
}

func mustParseOpts() error {

	// defaults
	opts.concurrency = 1
	opts.method = "GET"

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("missing argument: URL\n\n")
		usage(-1)
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			usage(0)
		case "-v", "--verbose":
			opts.verbose = true
		case "-c", "--concurrency":
			i++
			if i >= len(args) {
				fmt.Fprintf(os.Stderr, "missing concurrency argument\n\n")
				usage(-1)
			}
			c, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid concurrency argument: %s\n\n", err.Error())
				usage(-1)
			}
			if c <= 0 {
				fmt.Fprintf(os.Stderr, "invalid concurrency argument: %d\n\n", c)
				usage(-1)
			}
			opts.concurrency = c
		case "-t", "--time":
			i++
			if i >= len(args) {
				fmt.Fprintf(os.Stderr, "missing reps argument\n\n")
				usage(-1)
			}
			d, err := time.ParseDuration(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid time argument: %s\n\n", err.Error())
				usage(-1)
			}
			opts.period = d
		case "-X", "--request":
			i++
			if i >= len(args) {
				fmt.Fprintf(os.Stderr, "missing request method argument\n\n")
				usage(-1)
			}
			opts.method = args[i]
		default:
			u, err := url.Parse(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid url: %s\n\n", err.Error())
				usage(-1)
			}
			opts.url = u
		}
	}
	return nil
}

func main() {
	mustParseOpts()

	// make http.Request
	req, err := http.NewRequest(opts.method, opts.url.String(), nil)
	if err != nil {
		fmt.Printf("failed to generate request: %s\n", err.Error())
		os.Exit(-1)
	}

	wg := &sync.WaitGroup{}

	start := time.Now()

	// generate workers
	for id := 0; id < opts.concurrency; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

			cli := &http.Client{
				Transport: tr,
			}
			for {

				rstart := time.Now()
				b, err := submitRequest(cli, req, opts.verbose && id == 0)
				d := time.Since(rstart)

				// print result
				if opts.verbose && id == 0 {
					if err != nil {
						fmt.Printf("[%s] err: %s\n", d.String(), err.Error())
					} else {
						fmt.Printf("[%s] %s\n", d.String(), string(b))
					}
				}

				if time.Since(start) > opts.period {
					return
				}
			}
		}(id)
	}
	wg.Wait()

}

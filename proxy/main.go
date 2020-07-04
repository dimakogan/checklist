// Starter code for proxy from:
//   https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
)

const SAFEBROWSING_ENDPOINT = "safebrowsing.googleapis.com"

func handleHTTP(w http.ResponseWriter, req *http.Request) {

	// Case: Update request

	// Case: Hash search request

	// Default: Error

	log.Printf("%v", req.Host)
	log.Printf("%v", req.URL)
	log.Printf("%v", req.URL.Host)
	req.URL.Scheme = "https"
	req.Host = SAFEBROWSING_ENDPOINT
	req.URL.Host = req.Host
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("error: %v", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	log.Printf("header: %v", resp.Header)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {
	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleHTTP(w, r)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Fatal(server.ListenAndServe())
}

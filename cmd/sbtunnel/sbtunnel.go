package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

  . "checklist/safebrowsing"
)

const (
	pathUpdate = "/v4/threatListUpdates:fetch"
	pathFind   = "/v4/fullHashes:find"

	firefoxQueryRequestKey = "$req"
)

func unmarshalFetch(req *http.Request) (*FetchThreatListUpdatesRequest, int, error) {
	q := req.URL.Query()
	body, ok := q[firefoxQueryRequestKey]

	if !ok {
		return nil, 0, errors.New("Query string does not contain $req field")
	}

	if len(body) != 1 {
		return nil, 0, errors.New("not sure what to do here")
	}

	data, err := base64.URLEncoding.DecodeString(body[0])
	if err != nil {
		return nil, 0, err
	}

	fetchHash := &FetchThreatListUpdatesRequest{}
	err = proto.Unmarshal(data, fetchHash)
	return fetchHash, len(data), err
}

// From sbserver source
// unmarshal reads pbResp from req. The mime will either be JSON or ProtoBuf.
func unmarshalFind(req *http.Request) (*FindFullHashesRequest, int, error) {
	q := req.URL.Query()
	body, ok := q[firefoxQueryRequestKey]

	if !ok {
		return nil, 0, errors.New("Query string does not contain $req field")
	}

	if len(body) != 1 {
		return nil, 0, errors.New("not sure what to do here")
	}

	data, err := base64.URLEncoding.DecodeString(body[0])
	if err != nil {
		return nil, 0, err
	}

	findHash := &FindFullHashesRequest{}
	err = proto.Unmarshal(data, findHash)
	return findHash, len(data), err
}

func formatThreatEntrySet(e *ThreatEntrySet) string {
	var logStr string
	if e.CompressionType == CompressionType_RAW {
		if e.GetRawHashes() != nil {
			decodedRawHashes, err := base64.StdEncoding.DecodeString(string(e.GetRawHashes().GetRawHashes()))
			if err != nil {
				fmt.Println("decode RawHashes error:", err)
			}
			logStr += fmt.Sprintf("RawHashes added: (prefixSize: %d, hashesStrLen: %d), ", e.GetRawHashes().PrefixSize, len(decodedRawHashes))
		}
		if e.GetRawIndices() != nil {
			logStr += fmt.Sprintf("RawIndices removed: %d, ", len(e.GetRawIndices().Indices))
		}
	} else {
		if e.GetRiceHashes() != nil {
			logStr += fmt.Sprintf("RiceHashes added: %d, ", e.GetRiceHashes().GetNumEntries())
		}
		if e.GetRiceIndices() != nil {
			logStr += fmt.Sprintf("RiceIndices removed: %d, ", e.GetRiceIndices().GetNumEntries())
		}
	}

	return logStr
}

func handleFetch(req *http.Request, respBody []byte) {
	fetchReq, reqLen, err := unmarshalFetch(req)
	if err != nil {
		log.Printf("Can't unmarshal fetchRequest: %v", err)
		return
	}
	fetchResp := &FetchThreatListUpdatesResponse{}
	if err := proto.Unmarshal(respBody, fetchResp); err != nil {
		log.Printf("Can't unmarshal fetchResponse: %v", err)
		ioutil.WriteFile("FetchThreatListUpdatesResponse.bin", respBody, 0644)
		return
	}
	logStr := fmt.Sprintf("FETCH Request: Size: %d bytes, ", reqLen)
	for _, lu := range fetchReq.ListUpdateRequests {
		logStr += fmt.Sprintf("%s, ", lu.ThreatType.String())
	}
	log.Printf(logStr)

	logStr = fmt.Sprintf("FETCH Response: Size: %d bytes, ", len(respBody))
	for _, lu := range fetchResp.ListUpdateResponses {
		logStr += fmt.Sprintf("{%s, %s, %s, ", lu.ThreatType.String(), lu.ResponseType.String(), lu.ThreatEntryType.String())
		for _, add := range lu.Additions {
			logStr += formatThreatEntrySet(add)
		}
		for _, rem := range lu.Removals {
			logStr += formatThreatEntrySet(rem)
		}
		logStr += "}, "
	}
	log.Printf(logStr)
}

func handleFind(req *http.Request, respBody []byte) {
	findReq, reqLen, err := unmarshalFind(req)
	if err != nil {
		log.Printf("Can't unmarshal findRequest: %v", err)
		return
	}
	findResp := &FindFullHashesResponse{}
	if err := proto.Unmarshal(respBody, findResp); err != nil {
		log.Printf("Can't unmarshal findResponse: %v", err)
		return
	}

	logStr := fmt.Sprintf("FIND Request: Size: %d bytes, %d prefixes, ", reqLen, len(findReq.ThreatInfo.ThreatEntries))
	for _, t := range findReq.ThreatInfo.ThreatTypes {
		logStr += fmt.Sprintf("%s, ", t.String())
	}
	log.Printf(logStr)

	log.Printf("FIND Response: Size: %d bytes, %d matches", len(respBody), len(findResp.Matches))
}

func handleHTTP(req *http.Request, respBody []byte) {
	if strings.HasPrefix(req.URL.Path, pathUpdate) {
		handleFetch(req, respBody)
		return
	} else if strings.HasPrefix(req.URL.Path, pathFind) {
		handleFind(req, respBody)
		return
	}
}

type transport struct {
	http.RoundTripper
	host string
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	handleHTTP(req, b)
	return resp, nil
}

func main() {
	port := flag.Int("p", 8888, "Listening port")
	sbServer := flag.String("s", "https://safebrowsing.googleapis.com", "SafeBrowsing server URL")
	logFile := flag.String("log", "sbtunnel.log", "Log file")
	flag.Parse()

	f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	origin, _ := url.Parse(*sbServer)

	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = origin.Host
		req.Host = origin.Host
		req.Header.Del("Accept-Encoding")
	}

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: &transport{http.DefaultTransport, origin.Host}}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Starting listening on :%d\n", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}

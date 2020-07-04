// Starter code for proxy from:
//   https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

package main

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/safebrowsing"
)

const (
	// Name of environment variable that holds API key
	apiKeyEnvVariable = "SAFEBROWSING_API_KEY"

	pathUpdate = "/v4/threatListUpdates:fetch"
	pathFind   = "/v4/fullHashes:find"

	mimeProto = "application/x-protobuf"

	firefoxQueryRequestKey = "$req"
)

type proxyState struct {
	cfg        safebrowsing.Config
	localIndex *LocalIndex
}

func handleUpdate(w http.ResponseWriter, req *http.Request) {
	req.URL.Scheme = "https"
	req.Host = safebrowsing.DefaultServerURL
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

// From sbserver source
// unmarshal reads pbResp from req. The mime will either be JSON or ProtoBuf.
func unmarshal(req *http.Request) (*FindFullHashesRequest, error) {
	q := req.URL.Query()
	body, ok := q[firefoxQueryRequestKey]

	if !ok {
		return nil, errors.New("Query string does not contain $req field")
	}

	if len(body) != 1 {
		return nil, errors.New("not sure what to do here")
	}

	data, err := base64.URLEncoding.DecodeString(body[0])
	if err != nil {
		return nil, err
	}

	findHash := &FindFullHashesRequest{}
	err = proto.Unmarshal(data, findHash)
	return findHash, err
}

func handleFind(state *proxyState, w http.ResponseWriter, req *http.Request) {
	hashReq := new(FindFullHashesRequest)
	hashReq, err := unmarshal(req)
	if err != nil {
		log.Printf("Can't unmarshal: %v", err)
		return
	}

	threats := hashReq.ThreatInfo
	if threats == nil {
		log.Printf("No threats")
	} else {
		entries := hashReq.ThreatInfo.ThreatEntries
		for i, e := range entries {
			idx, _ := state.localIndex.GetIndex(hashPrefix(e.Hash))
			log.Printf("Hash[%v] = %v [index %v]", i, e.Hash, idx)
		}
	}
}

func handleHTTP(state *proxyState, w http.ResponseWriter, req *http.Request) {
	log.Printf("%v", req.Host)
	log.Printf("%v", req.URL)
	log.Printf("%v", req.URL.Host)

	if strings.HasPrefix(req.URL.Path, pathUpdate) {
		log.Printf("UPDATE")
		handleUpdate(w, req)
		return
	} else if strings.HasPrefix(req.URL.Path, pathFind) {
		log.Printf("FIND")
		handleFind(state, w, req)
		return
	}

	// Some other API call that we don't know how to handle.
	http.Error(w, "Not found", http.StatusNotFound)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func getConfig(dbFile string) safebrowsing.Config {
	key, isPresent := os.LookupEnv(apiKeyEnvVariable)
	if !isPresent {
		log.Fatalf("Could not find SafeBrowsing API key in environment variable '%v'", apiKeyEnvVariable)
	}

	var cfg safebrowsing.Config
	cfg.APIKey = key
	cfg.DBPath = dbFile
	cfg.Logger = os.Stderr
	cfg.UpdatePeriod = 5 * time.Minute

	return cfg
}

func main() {
	dbFile, err := ioutil.TempFile(os.TempDir(), "database-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	dbFile.Close()
	defer os.Remove(dbFile.Name())

	state := proxyState{
		cfg:        getConfig(dbFile.Name()),
		localIndex: NewLocalIndex(dbFile.Name())}
	defer state.localIndex.Close()

	sb, err := safebrowsing.NewSafeBrowser(state.cfg)

	if err != nil {
		log.Fatalf("Cannot create SafeBrowser: %v", err)
	}

	defer sb.Close()

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleHTTP(&state, w, r)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Fatal(server.ListenAndServe())
}

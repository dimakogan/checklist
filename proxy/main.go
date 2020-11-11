// Starter code for proxy from:
//   https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
  duration "github.com/golang/protobuf/ptypes/duration"
	"github.com/google/safebrowsing"
)

const (
	// Name of environment variable that holds API key
	apiKeyEnvVariable = "SAFEBROWSING_API_KEY"

	pathUpdate = "/v4/threatListUpdates:fetch"
	pathFind   = "/v4/fullHashes:find"

	mimeProto = "application/x-protobuf"

	firefoxQueryRequestKey = "$req"

  // in seconds
  minimiumWaitDuration = 3
)

type proxyState struct {
	cfg        safebrowsing.Config
	localIndex *LocalIndex
}

func unmarshalFetch(req *http.Request) (*FetchThreatListUpdatesRequest, error) {
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

	fetchHash := &FetchThreatListUpdatesRequest{}
	err = proto.Unmarshal(data, fetchHash)
  return fetchHash, err
}

// From sbserver source
// unmarshal reads pbResp from req. The mime will either be JSON or ProtoBuf.
func unmarshalFind(req *http.Request) (*FindFullHashesRequest, error) {
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

func marshal(resp http.ResponseWriter, pbResp proto.Message) error {
	resp.Header().Set("Content-Type", mimeProto)
	body, err := proto.Marshal(pbResp)

	if err != nil {
		return err
	}

	if _, err := resp.Write(body); err != nil {
		return err
	}

	return nil
}

func newThreatMatch(tt ThreatType, bytesIn []byte) *ThreatMatch {
	tm := new(ThreatMatch)
	tm.ThreatType = tt
	tm.PlatformType = PlatformType_ALL_PLATFORMS
	tm.ThreatEntryType = ThreatEntryType_URL

	tm.Threat = new(ThreatEntry)
	tm.Threat.Hash = bytesIn
	return tm

}

/*
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
}*/

func hashesToBigString(hashes []PartialHash) []byte {
  out := make([]byte, len(hashes) * PartialHashLen)
  for i,v := range hashes {
    copy(out[i*PartialHashLen:(i+1)*PartialHashLen], v[:])
  }
  return out
}

func computeSum(hashes []byte) []byte {
	hash := sha256.New()
	hash.Write([]byte(hashes))
  return hash.Sum(nil)
}

func handleFetch(state *proxyState, w http.ResponseWriter, req *http.Request) {
  /*
	upReq := new(FetchThreatListUpdatesRequest)
	upReq, err := unmarshalFetch(req)
	if err != nil {
		log.Printf("Can't unmarshal: %v", err)
		return
	}

  nLists := len(upReq.ListUpdateRequests)
  */

  hashes := hashesToBigString(getPartialHashes())

	resp := new(FetchThreatListUpdatesResponse)
	resp.ListUpdateResponses = make([]*FetchThreatListUpdatesResponse_ListUpdateResponse, 1)
  resp.MinimumWaitDuration = new(duration.Duration)
  resp.MinimumWaitDuration.Seconds = minimiumWaitDuration

  listUp := new(FetchThreatListUpdatesResponse_ListUpdateResponse)
  listUp.ThreatType = ThreatType_MALWARE
  listUp.ThreatEntryType = ThreatEntryType_URL
  listUp.PlatformType = PlatformType_ALL_PLATFORMS
  listUp.ResponseType = FetchThreatListUpdatesResponse_ListUpdateResponse_FULL_UPDATE
  listUp.Additions = make([]*ThreatEntrySet, 1)
  listUp.Additions[0] = new(ThreatEntrySet)
  listUp.Additions[0].RawHashes = new(RawHashes)
  listUp.Additions[0].RawHashes.PrefixSize = PartialHashLen
  listUp.Additions[0].RawHashes.RawHashes = hashes
  listUp.NewClientState = []byte("c3RhdGUgaXMgbm8gc3RhdGU=")

  listUp.Checksum = new(Checksum)
  listUp.Checksum.Sha256 = computeSum(listUp.Additions[0].RawHashes.RawHashes)

  resp.ListUpdateResponses[0] = listUp

	marshal(w, resp)
}

func handleFind(state *proxyState, w http.ResponseWriter, req *http.Request) {
	hashReq := new(FindFullHashesRequest)
	hashReq, err := unmarshalFind(req)
	if err != nil {
		log.Printf("Can't unmarshal: %v", err)
		return
	}

	threats := hashReq.ThreatInfo
	if threats == nil {
		log.Printf("No threats")
		return
	}

	entries := hashReq.ThreatInfo.ThreatEntries
	nThreats := len(entries)

	resp := new(FindFullHashesResponse)
	resp.Matches = make([]*ThreatMatch, nThreats)
	for i, e := range entries {
		idx, _ := state.localIndex.GetIndex(hashPrefix(e.Hash))
		h := queryForHash(idx)
		log.Printf("Hash[%v] = %x [index %v]", i, e.Hash, idx)
		resp.Matches[i] = newThreatMatch(hashReq.ThreatInfo.ThreatTypes[0], h)
		log.Printf("Returning hash: %x", h)
	}

	marshal(w, resp)
}



func handleHTTP(state *proxyState, w http.ResponseWriter, req *http.Request) {
	log.Printf("%v", req.Host)
	log.Printf("%v", req.URL)
	log.Printf("%v", req.URL.Host)

	if strings.HasPrefix(req.URL.Path, pathUpdate) {
		log.Printf("FETCH")
		handleFetch(state, w, req)
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

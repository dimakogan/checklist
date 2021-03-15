// Starter code for proxy from:
//   https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c

package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"

	"checklist/driver"
	"checklist/pir"
	. "checklist/safebrowsing"
	"checklist/updatable"
)

const (
	pathUpdate = "/v4/threatListUpdates:fetch"
	pathFind   = "/v4/fullHashes:find"

	mimeProto = "application/x-protobuf"

	firefoxQueryRequestKey = "$req"

	// in seconds
	minimiumWaitDuration = 60 * 60
)

type sbproxy struct {
	pirClient *updatable.Client
}

func NewSBProxy(serverAddr string) *sbproxy {
	addrs := strings.Split(serverAddr, ",")
	rpcLeft, err := driver.NewRpcProxy(addrs[0], true, true)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", addrs[0], err)
	}
	var rpcRight *driver.RpcProxy
	if len(addrs) > 1 {
		rpcRight, err = driver.NewRpcProxy(addrs[1], true, true)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", addrs[1], err)
		}
	} else {
		rpcRight = rpcLeft
	}

	client := updatable.NewClient(pir.RandSource(), pir.Punc, [2]updatable.UpdatableServer{rpcLeft, rpcRight})
	if err = client.Init(); err != nil {
		log.Fatalf("Failed to run PIR Init: %s\n", err)
	}
	log.Printf("PIR Init ok with %d keys\n", len(client.Keys()))
	return &sbproxy{pirClient: client}
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

func hashesToBigString(hashes []PartialHash) []byte {
	out := make([]byte, len(hashes)*PartialHashLen)
	for i, v := range hashes {
		copy(out[i*PartialHashLen:(i+1)*PartialHashLen], v[:])
	}
	return out
}

func (proxy *sbproxy) handleFetch(w http.ResponseWriter, req *http.Request) {
	upReq := new(FetchThreatListUpdatesRequest)
	upReq, err := unmarshalFetch(req)
	if err != nil {
		log.Printf("Can't unmarshal: %v", err)
		return
	}

	resp := new(FetchThreatListUpdatesResponse)
	resp.MinimumWaitDuration = new(duration.Duration)
	resp.MinimumWaitDuration.Seconds = minimiumWaitDuration

	nLists := len(upReq.ListUpdateRequests)
	resp.ListUpdateResponses = make([]*FetchThreatListUpdatesResponse_ListUpdateResponse, nLists)

	for i, v := range upReq.ListUpdateRequests {
		listUp := new(FetchThreatListUpdatesResponse_ListUpdateResponse)
		listUp.ThreatType = v.ThreatType
		listUp.ThreatEntryType = v.ThreatEntryType
		listUp.PlatformType = v.PlatformType
		listUp.ResponseType = FetchThreatListUpdatesResponse_ListUpdateResponse_FULL_UPDATE
		listUp.Additions = make([]*ThreatEntrySet, 1)
		listUp.Additions[0] = new(ThreatEntrySet)
		listUp.Additions[0].CompressionType = CompressionType_RICE
		listUp.Additions[0].RiceHashes, err = updatable.RiceEncodedHashes(proxy.pirClient.Keys())
		if err != nil {
			log.Printf("Can't Rice-encode hashes: %v", err)
			return
		}
		listUp.Additions[0].RawHashes = new(RawHashes)
		listUp.Additions[0].RawHashes.PrefixSize = 0
		listUp.Additions[0].RawHashes.RawHashes = make([]byte, 0)

		listUp.Removals = make([]*ThreatEntrySet, 0)

		// Just a garbage string.
		listUp.NewClientState = []byte("c3RhdGUgaXMgbm8gc3RhdGU=")

		// Use empty checksum since apparently Firefox doesn't inspect it
		// if the checksum is empty.
		listUp.Checksum = new(Checksum)
		listUp.Checksum.Sha256 = make([]byte, 0)

		resp.ListUpdateResponses[i] = listUp
	}

	marshal(w, resp)
}

func (proxy *sbproxy) handleFind(w http.ResponseWriter, req *http.Request) {
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
		prefixInt := binary.LittleEndian.Uint32(e.Hash)
		log.Printf("Looking for hash prefix = %x", prefixInt)
		h, err := proxy.pirClient.Read(prefixInt)
		if err != nil {
			log.Printf("PIR query failed: %s\n", err)
			continue
		}
		resp.Matches[i] = newThreatMatch(hashReq.ThreatInfo.ThreatTypes[0], h)
		log.Printf("Returning hash: %x", h)
	}

	marshal(w, resp)
}

func (proxy *sbproxy) handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%v", req.Host)
	log.Printf("%v", req.URL)
	log.Printf("%v", req.URL.Host)

	if strings.HasPrefix(req.URL.Path, pathUpdate) {
		log.Printf("FETCH")
		proxy.handleFetch(w, req)
		return
	} else if strings.HasPrefix(req.URL.Path, pathFind) {
		log.Printf("FIND")
		proxy.handleFind(w, req)
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

func main() {
	var serverAddr string
	flag.StringVar(&serverAddr, "serverAddr", ":12345", "<HOSTNAME>:<PORT> of one or two comma-separated PIR servers")
	flag.Parse()

	proxy := NewSBProxy(serverAddr)

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.handleHTTP(w, r)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Printf("Listening on %s...", server.Addr)
	log.Fatal(server.ListenAndServe())
}

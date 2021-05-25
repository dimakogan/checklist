[![Actions Status](https://github.com/dimakogan/checklist/actions/workflows/main.yml/badge.svg)](https://github.com/dimakogan/checklist/actions/workflows/main.yml)

# Checklist: Private Blocklist Lookups

*This code accompanies the [Checklist paper](https://eprint.iacr.org/2021/345.pdf) by [Dmitry Kogan](https://cs.stanford.edu/~dkogan/) and [Henry Corrigan-Gibbs](https://people.csail.mit.edu/henrycg/).*

Checklist is a system for privacy-preserving blocklist lookups. That is: a pair of servers holds a set *B* of blocklisted strings, a client holds a private string *s*, and the client wants to learn whether *s* is in the blocklist *B* without revealing its private string *s* to the servers.

The technical components of Checklist are:

* a new two-server private-information-retrieval protocol (the `pir/` directory),
* a technique that extends preprocessing offline/online PIR schemes to support database updates (see the `updatable/` directory), and
* a Safe-Browsing service proxy that allows Firefox to perform Safe Browsing API lookups via Checklist (see the `cmd/sbproxy/` and `cmd/rpcserver/` directories).

### Code organization 

The directories in this repository are:

| **Core PIR library** ||
| :--- | :---|
| [pir/](pir/) | Core PIR protocol code |
| [psetggm/](psetggm/) | Optimized C++ implementation of puncturable-set primitive|
| **Extension: Database updates** | |
|[updatable/](updatable/) | Implementation of offline/online PIR with database updates|
| **Networking and benchmarking** | |
| [driver/](driver/) |Wrapper code for benchmarks |
| [rpc/](rpc/) | RPC over HTTPS |
| **Examples and applications** | |
| [example/](example/) | Example of how to invoke our basic PIR library |
| [cmd/rpc_server](cmd/rpc_server/) | Checklist server executable |
| [cmd/sbproxy](cmd/sbproxy/) | Code to proxy Firefox SafeBrowsing requests through Checklist |


### PIR library

Our implementation supports the following PIR protocols, which we implement in the `pir/` directory. In the bulleted list below, λ ≈ 128 is the security parameter and n is the number of rows in the database. We ignore leading constants.

* `pir.Punc` - Checklist's new two-server offline/online PIR scheme. The PIR scheme has offline server time λn, offline communication λn^{1/2}, online server time n^{1/2}, and online communication λlog(n).
* `pir.Matrix` - A simple two-server PIR scheme based on the original PIR paper of [Chor, Goldreich, Kushilevitz, and Sudan](http://www.wisdom.weizmann.ac.il/~oded/PSX/pir2.pdf). The PIR scheme has offline server time 0, offline communication 0, online server time n, and online communication n^{1/2}.
* `pir.DPF` - A DPF-based two-server PIR, based on the "[Function Secret Sharing](https://eprint.iacr.org/2018/707)" work of Boyle, Gilboa, and Ishai. The PIR scheme has offline server time 0, offline communication 0, online server time n, and online communication λlog(n).
* `pir.NonPrivate` - Fetch a database record with no privacy.

### Safe Browsing proxy for Firefox

To try running Checklist's Safe Browsing proxy for Firefox, follow these steps, each in a **separate terminal**, starting from the repository root.

**1. Build the project**

```
$ go build ./...
```

**2. Run Checklist servers**

```
# Run two Checklist servers in the background
$ go run ./cmd/rpc_server -f safebrowsing/evil_urls.txt -p 8800 &
$ go run ./cmd/rpc_server -f safebrowsing/evil_urls.txt -p 8801
```

**3. Run the local Safe Browsing proxy**

```
$ go run ./cmd/sbproxy -serverAddr=localhost:8800,localhost:8801   
# Listens on localhost:8888
```

**4. Run Firefox with a modified profile** 

The directory `safebrowsing/ff-profile` contains a Firefox profile that's configured to make Safe Browsing API requests to the proxy at `localhost:8888`.

```
$ cd safebrowsing
$ ./run_browser.sh
```

When you open Firefox, you should see some activity on the proxy and PIR servers. Test the system is working by navigating to [https://en.wikipedia.org/wiki/Main_Page](https://en.wikipedia.org/wiki/Main_Page), which we added as a test URL in [safebrowsing/evil_urls.txt](safebrowsing/evil_urls.txt)

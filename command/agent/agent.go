package agent

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/hashicorp/go-discover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/umbracle/minimal/network/discovery"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/minimal/blockchain"
	"github.com/umbracle/minimal/blockchain/storage"
	"github.com/umbracle/minimal/chain"
	"github.com/umbracle/minimal/consensus/ethash"
	"github.com/umbracle/minimal/network"
	"github.com/umbracle/minimal/network/rlpx"
	"github.com/umbracle/minimal/protocol"
	"github.com/umbracle/minimal/protocol/ethereum"
	"github.com/umbracle/minimal/syncer"

	discoveryConsul "github.com/umbracle/minimal/network/discovery/consul"
	discoveryDevP2P "github.com/umbracle/minimal/network/discovery/devp2p"
)

// Agent is a long running daemon that is used to run
// the ethereum client
type Agent struct {
	logger *log.Logger
	config *Config

	server   *network.Server
	discover *discover.Discover
	syncer   *syncer.Syncer
}

func NewAgent(logger *log.Logger, config *Config) *Agent {
	return &Agent{logger: logger, config: config}
}

// Start starts the agent
func (a *Agent) Start() error {
	a.startTelemetry()

	chain, err := chain.ImportFromName(a.config.Chain)
	if err != nil {
		return fmt.Errorf("Failed to load chain %s: %v", a.config.Chain, err)
	}

	// Load private key from memory (TODO, do it from a file)
	privateKey := "b4c65ef6b82e96fb5f26dc10a79c929985217c078584721e9157c238d1690b22"

	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	// Start server
	serverConfig := network.DefaultConfig()
	serverConfig.BindAddress = a.config.BindAddr
	serverConfig.BindPort = a.config.BindPort
	serverConfig.Bootnodes = chain.Bootnodes

	serverConfig.DiscoveryBackends = map[string]discovery.Factory{
		"devp2p": discoveryDevP2P.Factory,
		"consul": discoveryConsul.Factory,
	}

	a.server = network.NewServer("minimal", key, serverConfig, a.logger)

	// Load consensus engine (TODO, configurable)
	consensus := ethash.NewEthHash(chain.Params, true)

	// blockchain storage
	storage, err := storage.NewLevelDBStorage(a.config.DataDir, nil)
	if err != nil {
		panic(err)
	}

	// blockchain object
	blockchain := blockchain.NewBlockchain(storage, consensus, chain.Params)
	if err := blockchain.WriteGenesis(chain.Genesis); err != nil {
		panic(err)
	}

	// Start syncer
	syncerConfig := syncer.DefaultConfig()
	syncerConfig.NumWorkers = 1

	// TODO, get network id from chain object
	a.syncer, err = syncer.NewSyncer(1, blockchain, syncerConfig)
	if err != nil {
		panic(err)
	}

	// register protocols
	callback := func(conn rlpx.Conn, peer *network.Peer) protocol.Handler {
		return ethereum.NewEthereumProtocol(conn, peer, a.syncer.GetStatus, blockchain)
	}
	a.server.RegisterProtocol(protocol.ETH63, callback)

	// Start network server work after all the protocols have been registered
	a.server.Schedule()

	// Start the syncer
	go a.syncer.Run()

	// Pipe new added nodes into syncer
	go func() {
		for {
			select {
			case evnt := <-a.server.EventCh:
				if evnt.Type == network.NodeJoin {
					a.syncer.AddNode(evnt.Peer)
				}
			}
		}
	}()

	return nil
}

// TODO, start the api service and connect the internal api with metrics
func (a *Agent) startTelemetry() {
	memSink := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(memSink)

	metricsConf := metrics.DefaultConfig("minimal")
	metricsConf.EnableHostnameLabel = false
	metricsConf.HostName = ""

	var sinks metrics.FanoutSink

	prom, err := prometheus.NewPrometheusSink()
	if err != nil {
		panic(err)
	}

	sinks = append(sinks, prom)
	sinks = append(sinks, memSink)

	metrics.NewGlobal(metricsConf, sinks)

	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(resp http.ResponseWriter, req *http.Request) {
		handler := promhttp.Handler()
		handler.ServeHTTP(resp, req)
	})

	go http.Serve(l, mux)
}

// Close stops the agent
func (a *Agent) Close() {
	// TODO, close syncer first
	a.server.Close()
}

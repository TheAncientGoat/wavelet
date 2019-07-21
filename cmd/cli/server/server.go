// Package server half-assedly provides functions and methods for a
// Wavelet server to interact with the TUI.
package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher"
	"github.com/perlin-network/noise/handshake"
	"github.com/perlin-network/noise/nat"
	"github.com/perlin-network/noise/skademlia"
	"github.com/perlin-network/wavelet"
	"github.com/perlin-network/wavelet/api"
	"github.com/perlin-network/wavelet/cmd/cli/tui/logger"
	"github.com/perlin-network/wavelet/internal/snappy"
	"github.com/perlin-network/wavelet/store"
	"github.com/perlin-network/wavelet/sys"
	"google.golang.org/grpc"
)

// Config contains the server configs
type Config struct {
	NAT      bool
	Host     string
	Port     uint
	Wallet   string
	Genesis  string
	APIPort  uint
	Peers    []string
	Database string
}

// Server provides a Wavelet server
type Server struct {
	logger *logger.Logger
	cfg    Config

	listener   net.Listener
	address    string
	tcpAddress *net.TCPAddr

	db store.KV

	Client *skademlia.Client
	Ledger *wavelet.Ledger
	Keys   *skademlia.Keypair

	History HistoryStore
}

// New creates a new server. `log' is needed.
func New(cfg Config, log *logger.Logger) (*Server, error) {
	s := &Server{
		logger: log,
		cfg:    cfg,
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	s.listener = l
	s.tcpAddress = l.Addr().(*net.TCPAddr)

	s.address = net.JoinHostPort(
		cfg.Host,
		strconv.Itoa(s.tcpAddress.Port),
	)

	if cfg.NAT {
		if len(cfg.Peers) > 1 {
			if err := nat.NewPMP().AddMapping(
				"tcp",
				uint16(s.tcpAddress.Port),
				uint16(s.tcpAddress.Port),
				30*time.Minute,
			); err != nil {
				return nil, err
			}
		}

		resp, err := http.Get("http://myexternalip.com/raw")
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		ip, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		s.address = net.JoinHostPort(
			string(ip),
			strconv.Itoa(s.tcpAddress.Port),
		)
	}

	s.logger.Level(logger.WithInfo("Listening for peers").
		F("address", s.address))

	// Read the keys from file or private key
	k, err := s.readKeys(cfg.Wallet)
	if err != nil {
		return nil, err
	}

	s.Keys = k

	// Set up a new S/Kademlia client
	s.Client = skademlia.NewClient(
		s.address, k,
		skademlia.WithC1(sys.SKademliaC1),
		skademlia.WithC2(sys.SKademliaC2),
		skademlia.WithDialOptions(
			grpc.WithDefaultCallOptions(grpc.UseCompressor(snappy.Name)),
		),
	)

	// Initiate events

	s.Client.SetCredentials(noise.NewCredentials(
		s.address, handshake.NewECDH(), cipher.NewAEAD(), s.Client.Protocol(),
	))

	s.Client.OnPeerJoin(func(conn *grpc.ClientConn, id *skademlia.ID) {
		pub := fmt.Sprintf("%x", id.PublicKey())
		s.History.add(pub, HistoryEntryAccount)

		s.logger.Level(logger.WithInfo("Peer joined").
			F("public_key", "%s", pub).
			F("address", id.Address()))
	})

	s.Client.OnPeerLeave(func(conn *grpc.ClientConn, id *skademlia.ID) {
		pub := fmt.Sprintf("%x", id.PublicKey())
		s.History.add(pub, HistoryEntryAccount)

		s.logger.Level(logger.WithInfo("Peer left").
			F("public_key", "%s", pub).
			F("address", id.Address()))
	})

	// Initialize the key-value store
	if cfg.Database != "" {
		kv, err := store.NewLevelDB(s.cfg.Database)
		if err != nil {
			// s.logger.Level(logger.WithError(err).
			// Wrap("Failed to create/open database").
			// F("location", s.cfg.Database))

			return nil, err
		}

		s.db = kv
	} else {
		// Make one in memory instead
		s.db = store.NewInmem()
	}

	// Create a new ledger
	s.Ledger = wavelet.NewLedger(s.db, s.Client)

	return s, nil
}

// Start runs the listeners. This does not block.
func (s *Server) Start() {
	go func() {
		server := s.Client.Listen()

		wavelet.RegisterWaveletServer(server, s.Ledger.Protocol())

		if err := server.Serve(s.listener); err != nil {
			s.logger.Level(logger.WithError(err).
				Wrap("An error occured in the server"))
		}
	}()

	for _, addr := range s.cfg.Peers {
		if _, err := s.Client.Dial(addr); err != nil {
			s.logger.Level(logger.WithError(err).
				Wrap("Error dialing").
				F("address", addr))
		}
	}

	if peers := s.Client.Bootstrap(); len(peers) > 0 {
		lvl := logger.WithSuccess(fmt.Sprintf(
			"Bootstrapped with %d peer(s)", len(peers),
		))

		for i, peer := range peers {
			_i := strconv.Itoa(i)

			lvl.F(_i+"_address", peer.Address())
			lvl.F(_i+"_pubkey", "%x", peer.PublicKey())
		}

		s.logger.Level(lvl)
	}

	if s.cfg.APIPort > 0 {
		go api.New().StartHTTP(int(s.cfg.APIPort), s.Client, s.Ledger, s.Keys)
	}
}

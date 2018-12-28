package rlpx

import (
	"crypto/ecdsa"
	"errors"
	"net"

	"github.com/ethereum/go-ethereum/p2p/discv5"
)

// A Config structure is used to configure a Rlpx server.
type Config struct {
	prv  *ecdsa.PrivateKey
	pub  *ecdsa.PublicKey
	info *Info
}

// Server returns a new Rlpx server side connection
func Server(conn net.Conn, prv *ecdsa.PrivateKey, info *Info) *Connection {
	return &Connection{conn: conn, prv: prv, localInfo: info}
}

// Client returns a new Rlpx client side connection
func Client(conn net.Conn, prv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, info *Info) *Connection {
	return &Connection{conn: conn, prv: prv, pub: pub, localInfo: info, isClient: true}
}

// Listener implements a network listener for Rlpx sessions.
type Listener struct {
	net.Listener
	config *Config
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (*Connection, error) {
	rawConn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := Server(rawConn, l.config.prv, l.config.info)
	if err := conn.Handshake(); err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}

// NewListener creates a Listener which accepts Rlpx connections
func NewListener(inner net.Listener, config *Config) *Listener {
	l := new(Listener)
	l.Listener = inner
	l.config = config
	return l
}

// Listen creates a Rlpx listener accepting connections on the
// given network address using net.Listen.
func Listen(network, laddr string, config *Config) (*Listener, error) {
	if config == nil || (config.prv == nil) {
		return nil, errors.New("rlpx: private key not set")
	}
	l, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return NewListener(l, config), nil
}

// DialWithDialer connects to the given network address using dialer.Dial and
// then initiates a Rlpx handshake
func DialWithDialer(dialer *net.Dialer, network, addr string, config *Config) (*Connection, error) {
	rawConn, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	conn := Client(rawConn, config.prv, config.pub, config.info)
	if err := conn.Handshake(); err != nil {
		rawConn.Close()
		return nil, err
	}
	return conn, nil
}

// Dial connects to the given network address using net.Dial
// and then initiates a Rlpx handshake.
func Dial(network, addr string, config *Config) (*Connection, error) {
	return DialWithDialer(new(net.Dialer), network, addr, config)
}

// DialEnode connects to the given enode address using net.Dial
// and then initiates a Rlpx handshake.
func DialEnode(network, addr string, config *Config) (*Connection, error) {
	enode, err := discv5.ParseNode(addr)
	if err != nil {
		return nil, err
	}
	pub, err := enode.ID.Pubkey()
	if err != nil {
		return nil, err
	}
	tcpAddr := net.TCPAddr{IP: enode.IP, Port: int(enode.TCP)}
	return DialWithDialer(new(net.Dialer), network, tcpAddr.String(), &Config{pub: pub, prv: config.prv})
}

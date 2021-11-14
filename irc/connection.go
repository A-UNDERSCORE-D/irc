package irc

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"awesome-dragon.science/go/irc/irc/isupport"
	"awesome-dragon.science/go/irc/irc/numerics"
	"github.com/ergochat/irc-go/ircmsg"
)

// ConnectionConfig contains all the configuration options used by Server
type ConnectionConfig struct {
	Host                  string // Hostname of the target server
	Port                  string // Server port
	TLS                   bool   // Use TLS
	InsecureSkipVerifyTLS bool   // Skip verifying TLS Certificates
	TLSCertPath           string
	TLSKeyPath            string
	RawLog                bool // Log raw messages
	DebugLog              bool // Log additional debug messages
}

// Connection implements the barebones required to make a connection to an IRC server.
//
// It expects that you do EVERYTHING yourself. It simply is a nice frontend for the socket.
type Connection struct {
	config *ConnectionConfig

	doneChan chan struct{}

	conn          net.Conn
	connectionCtx context.Context
	cancelConnCtx context.CancelFunc
	lineChan      chan *ircmsg.Message // Incoming lines
	writeMutex    sync.Mutex           // Protects the write socket

	log *log.Logger

	ISupport *isupport.ISupport
}

// NewConnection creates a new Server instance ready for use
func NewConnection(config *ConnectionConfig) *Connection {
	return &Connection{
		config:   config,
		doneChan: make(chan struct{}),
		lineChan: make(chan *ircmsg.Message),
		log:      log.Default(),
		ISupport: isupport.New(),
	}
}

// NewSimpleServer is a nice wrapper that creates a ServerConfig for you
func NewSimpleServer(host, port string, useTLS bool) *Connection {
	return NewConnection(&ConnectionConfig{Host: host, Port: port, TLS: useTLS, DebugLog: true, RawLog: true})
}

func (s *Connection) debugLog(v ...interface{}) {
	if s.config.DebugLog {
		s.log.Print(v...)
	}
}

func (s *Connection) debugLogf(format string, v ...interface{}) {
	if s.config.DebugLog {
		s.log.Printf(format, v...)
	}
}

// Connect connects the Server instance to IRC. It does NOT block.
func (s *Connection) Connect(ctx context.Context) error {
	s.debugLog("Opening connection")

	conn, err := s.openConn(ctx)
	if err != nil {
		return fmt.Errorf("could not open connection: %w", err)
	}

	s.conn = conn
	mainCtx, mainCancel := context.WithCancel(context.Background())

	s.connectionCtx = mainCtx
	s.cancelConnCtx = mainCancel

	// These are being used as signals
	readCtx, readCancel := context.WithCancel(mainCtx)

	_ = readCancel

	go s.readLoop(readCtx) //nolint:contextcheck // Its because we create a NEW one that's unrelated to the Connect one

	return nil
}

func (s *Connection) openConn(ctx context.Context) (net.Conn, error) {
	var (
		conn     net.Conn
		err      error
		hostPort = net.JoinHostPort(s.config.Host, s.config.Port)
	)

	s.debugLogf("Opening connection to %q...", hostPort)

	if s.config.TLS {
		dialer := &tls.Dialer{}
		dialer.Config = &tls.Config{InsecureSkipVerify: s.config.InsecureSkipVerifyTLS} //nolint:gosec // Its intentional

		if s.config.TLSCertPath != "" && s.config.TLSKeyPath != "" {
			res, err := tls.LoadX509KeyPair(s.config.TLSCertPath, s.config.TLSKeyPath)
			if err != nil {
				return nil, fmt.Errorf("could not load keypair: %w", err)
			}

			dialer.Config.Certificates = append(dialer.Config.Certificates, res)
		}

		conn, err = dialer.DialContext(ctx, "tcp", hostPort)
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(s.config.Host, s.config.Port))
	}

	if err != nil {
		return nil, fmt.Errorf("could not dial: %w", err)
	}

	return conn, nil
}

func (s *Connection) readLoop(ctx context.Context) {
	reader := bufio.NewReader(s.conn)

outer:
	for {
		select {
		case <-ctx.Done():
			break outer

		default:
		}

		data, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			s.log.Printf("Unexpected error from conn.Read: %s", err)

			break
		}

		msg, err := ircmsg.ParseLine(data)
		if err != nil {
			s.log.Printf("got an invalid IRC Line: %q -> %s", data, err)

			continue
		}

		if s.config.RawLog {
			s.log.Printf("[>>] %s", data)
		}

		s.onLine(&msg)
	}

	close(s.lineChan)
	s.cancelConnCtx()
}

func (s *Connection) onLine(msg *ircmsg.Message) {
	switch msg.Command {
	case numerics.RPL_ISUPPORT:
		s.ISupport.Parse(msg)
	case numerics.RPL_MYINFO:
	}

	s.lineChan <- msg
}

func (s *Connection) Write(b []byte) (int, error) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if s.config.RawLog {
		s.log.Printf("[<<] %s", b)
	}

	return s.conn.Write(b) //nolint:wrapcheck // Its the usual.
}

// WriteLine constructs an ircmsg.Message and sends it to the server
func (s *Connection) WriteLine(command string, args ...string) error {
	msg := ircmsg.MakeMessage(nil, "", command, args...)

	bytes, err := msg.LineBytes()
	if err != nil {
		return fmt.Errorf("could not create IRC line: %w", err)
	}

	_, err = s.Write(bytes)

	return err
}

// WriteString implements io.StringWriter
func (s *Connection) WriteString(m string) (int, error) { return s.Write([]byte(m)) } //nolint:gocritic // ... No

// LineChan returns a read only channel that will have messages from the server
// sent to it
func (s *Connection) LineChan() <-chan *ircmsg.Message { return s.lineChan }

// Done returns a channel that is closed when the connection is closed.
func (s *Connection) Done() <-chan struct{} { return s.connectionCtx.Done() }

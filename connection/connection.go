package connection

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"awesome-dragon.science/go/irc/isupport"
	"awesome-dragon.science/go/irc/numerics"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("irc") //nolint:gochecknoglobals // Its the logger.

// Config contains all the configuration options used by Server
type Config struct {
	Host                  string // Hostname of the target server
	Port                  string // Server port
	TLS                   bool   // Use TLS
	InsecureSkipVerifyTLS bool   // Skip verifying TLS Certificates
	TLSCertPath           string
	TLSKeyPath            string
	RawLog                bool // Log raw messages
}

// Connection implements the barebones required to make a connection to an IRC server.
//
// It expects that you do EVERYTHING yourself. It simply is a nice frontend for the socket.
type Connection struct {
	config *Config

	conn          net.Conn
	connectionCtx context.Context // nolint:containedctx // Used to hold onto tne entire connection
	cancelConnCtx context.CancelFunc
	lineChan      chan *ircmsg.Message // Incoming lines
	writeMutex    sync.Mutex           // Protects the write socket

	ISupport *isupport.ISupport
}

// NewConnection creates a new Server instance ready for use
func NewConnection(config *Config) *Connection {
	return &Connection{
		config:   config,
		lineChan: make(chan *ircmsg.Message),
		ISupport: isupport.New(),
	}
}

// NewSimpleServer is a nice wrapper that creates a ServerConfig for you
func NewSimpleServer(host, port string, useTLS bool) *Connection {
	return NewConnection(&Config{Host: host, Port: port, TLS: useTLS, RawLog: true})
}

// Connect connects the Server instance to IRC. It does NOT block.
func (s *Connection) Connect(ctx context.Context) error {
	connContext, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err := s.openConn(connContext)
	if err != nil {
		return fmt.Errorf("could not open connection: %w", err)
	}

	s.conn = conn
	mainCtx, mainCancel := context.WithCancel(ctx)

	s.connectionCtx = mainCtx
	s.cancelConnCtx = mainCancel

	// These are being used as signals
	readCtx, readCancel := context.WithCancel(mainCtx)

	_ = readCancel

	go s.readLoop(readCtx)

	return nil
}

func (s *Connection) openConn(ctx context.Context) (net.Conn, error) {
	var (
		conn     net.Conn
		err      error
		hostPort = net.JoinHostPort(s.config.Host, s.config.Port)
	)

	log.Debugf("Opening connection to %q...", hostPort)

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

			log.Warningf("Unexpected error from conn.Read: %s", err)

			break
		}

		msg, err := ircmsg.ParseLine(data)
		if err != nil {
			log.Warningf("got an invalid IRC Line: %q -> %s", data, err)

			continue
		}

		if s.config.RawLog {
			log.Infof("[>>] %s", data)
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
		log.Infof("[<<] %s", b)
	}

	n, err := s.conn.Write(b)
	if err != nil {
		return n, fmt.Errorf("Connection.Write: %w", err)
	}

	return n, nil
}

// WriteLine constructs an ircmsg.Message and sends it to the server
func (s *Connection) WriteLine(command string, args ...string) error {
	msg := ircmsg.MakeMessage(nil, "", command, args...)

	bytes, err := msg.LineBytes()
	if err != nil {
		return fmt.Errorf("could not create IRC line: %w", err)
	}

	_, err = s.Write(bytes)
	if err != nil {
		return fmt.Errorf("WriteLine: Could not send line: %w", err)
	}

	return nil
}

// WriteString implements io.StringWriter
func (s *Connection) WriteString(m string) (int, error) { return s.Write([]byte(m)) } //nolint:gocritic // ... No

// LineChan returns a read only channel that will have messages from the server
// sent to it
func (s *Connection) LineChan() <-chan *ircmsg.Message { return s.lineChan }

// Done returns a channel that is closed when the connection is closed.
func (s *Connection) Done() <-chan struct{} { return s.connectionCtx.Done() }

// Stop stops the connection to IRC
func (s *Connection) Stop(msg string) {
	if err := s.WriteLine("QUIT", msg); err != nil {
		log.Infof("Failed to write quit while exiting: %s", err)
	}

	select {
	case <-time.After(time.Second * 2):
		s.cancelConnCtx()
	case <-s.connectionCtx.Done():
	}
}

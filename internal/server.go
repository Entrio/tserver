package internal

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
)

var (
	instance *Server
)

type (
	Server struct {
		name     string
		rwm      sync.RWMutex
		clients  map[uuid.UUID]*Client
		addr     *net.TCPAddr
		listener *net.TCPListener
	}
)

func (s *Server) GetName() string {
	return s.name
}

func NewServer(ipep string) *Server {
	if instance != nil {
		log.Info().Msg("Returning existing server instance")
		return instance
	}

	instance := &Server{
		clients: make(map[uuid.UUID]*Client),
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ipep)

	if err != nil {
		log.Err(err).Msg("Failed to parse IP Endpoint")
		return nil
	}

	instance.addr = tcpAddr

	log.Info().Msg("New server instance created")
	return instance
}

// Start the server on provided TCP Endpoint
func (s *Server) Start() (err error) {
	log.Info().Msg("Attempting to start the server")
	s.listener, err = net.ListenTCP("tcp", s.addr)
	if err != nil {
		log.Err(err).Msg("Failed to start the server")
		return
	}

	defer s.listener.Close()
	log.Info().Str("IPEP", s.addr.String()).Msg("Server started")
	s.listenForConnections()
	return nil
}

// addCClient adds a new client to the map and returns the instance of the new client
func (s *Server) addCClient(cc net.Conn) *Client {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	c := &Client{
		ID:         uuid.New(),
		Connection: cc,
	}
	s.clients[c.ID] = c
	log.Info().Str("ID", c.ID.String()).Msg("New client connected")
	go s.handleData(c)
	return c
}

// removeClient removes the client from the list of clients by their client instance
func (s *Server) removeClient(c *Client) error {
	return nil
}

// removeClientByID removes a client from the client list by their UID
func (s *Server) removeClientByID(uid string) error {
	return nil
}

// BroadcastAll send message to all connected clients
func (s *Server) BroadcastAll(msg []byte) error {
	return nil
}

// BroadcastClient send a message directly to client
func (s *Server) BroadcastClient(uid string, msg []byte) error {
	return nil
}

// listenForConnections loops forever and waits for incoming connections
func (s *Server) listenForConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Warn().Msg("Failed to accept tcp connection")
			log.Err(err).Msg("Error:")
			continue
		}
		s.addCClient(conn)
	}
}

func (s *Server) handleData(c *Client) {
	log.Info().Str("client", c.ID.String()).Msg("Starting a goroutine to handle data for ")
	recBuff := make([]byte, 1024, 1204) // 1mb buffer, this might become a problem when there are many clients. Should agree on max packet size beforehand
	/*
		packet := Packet{
			buffer: make([]byte, 0),
			cursor: 0,
			conn:   &c.Connection,
		}*/

	for {
		cLen, err := c.Connection.Read(recBuff)

		if err != nil {
			log.Err(err).Msg("Error on client read")
			c.Connection.Close()
			s.removeClient(c)
			return
		}

		log.Info().Int("size", cLen).Msg("Got some data")
		fmt.Println(string(recBuff))
	}
}

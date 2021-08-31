package internal

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strings"
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
		input    *bufio.Reader
		exitChan chan struct{}
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

	instance = &Server{
		clients:  make(map[uuid.UUID]*Client),
		name:     "FIRE PHOENIX testing server",
		input:    bufio.NewReader(os.Stdin),
		exitChan: make(chan struct{}),
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ipep)

	if err != nil {
		log.Err(err).Msg("Failed to parse IP Endpoint")
		return nil
	}

	instance.addr = tcpAddr
	go instance.listenForInput()

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
	go s.listenForConnections()
	<-s.exitChan
	log.Info().Msg("Server terminated")
	return nil
}

func (s *Server) listenForInput() {
	for {
		select {
		case <-s.exitChan:
			return
		default:
			data, err := s.input.ReadBytes('\n')
			if len(data) <= 1 {
				continue
			}
			if err != nil {
				log.Err(err).Msg("Failed to parse input")
			}
			data = data[:len(data)-1] // remove the \n

			var payload string
			if string(data[:1]) == "/" {
				// we are issuing a command
				payload = string(data[1:])

				// is there a space (we need to get the args after the command)
				spacePos := strings.Index(payload, " ")
				var cmd string
				if spacePos != -1 {
					// there are spaces, meaning there are args
					cmd = string(data[1 : spacePos+1])
					args := strings.Split(string(data[spacePos+2:]), " ") // compensate the space and the /
					s.processCommands(cmd, args...)
				} else {
					cmd = string(data[1:])
					s.processCommands(cmd)
				}

			} else {
				payload = string(data)

				// just a msg to everyone
				if err := s.BroadcastAll([]byte(fmt.Sprintf("[SERVER BROADCAST] %s\n", payload))); err != nil {
					log.Err(err).Msg("Failed to broadcast to all")
				}
			}
		}

	}
}

func (s *Server) processCommands(cmd string, args ...string) {
	log.Info().Str("cmd", cmd).Msg("Processing command")
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "quit": // Exit
		s.exitChan <- struct{}{}
		break
	case "w": // Send a message to a specific user
		if len(args) <= 1 {
			log.Warn().Msg("Please specify a message")
			break
		}

		destID, err := uuid.Parse(args[0])
		if err != nil {
			log.Err(err).Msg("failed to parse client ID")
			break
		}

		msg := fmt.Sprintf("%s\n", strings.Join(args[1:], " "))
		if err := s.BroadcastClient(destID, []byte(msg)); err != nil {
			log.Err(err).Msg("failed to broadcast message to client")
		}

		break
	}

	if len(args) > 0 {
		fmt.Println(args)
	}
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
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	for _, c := range s.clients {
		_, err := c.Connection.Write(msg)
		if err != nil {
			log.Warn().Err(err).Str("id", c.ID.String()).Msg("Failed to write to client socket")
		}
	}

	return nil
}

// BroadcastClient send a message directly to client
func (s *Server) BroadcastClient(uid uuid.UUID, msg []byte) error {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	if client, ok := s.clients[uid]; ok {
		_, werr := client.Connection.Write(msg)
		if werr != nil {
			log.Err(werr).Msg("Failed to write to client socket")
		}
	} else {
		log.Warn().Str("dst", uid.String()).Msg("Destination ID was not found")
	}

	return nil
}

// listenForConnections loops forever and waits for incoming connections
func (s *Server) listenForConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Warn().Msg("Failed to accept tcp connection")
			log.Err(err).Msg("Error:")
			break
		}
		s.addCClient(conn)
	}
}

func (s *Server) handleData(c *Client) {
	c.Connection.Write([]byte(fmt.Sprintf("Welcome to %s, your ID is %s\n", s.name, c.ID.String())))
	log.Info().Str("client", c.ID.String()).Msg("Starting a goroutine to handle data for ")
	recBuff := make([]byte, 1024*1024) // 1mb buffer size, this might cause an issue when many clients are connected

	packet := &Packet{
		buffer: make([]byte, 0),
		cursor: 0,
		client: c,
	}

	for {
		cLen, err := c.Connection.Read(recBuff)

		if err != nil {
			log.Err(err).Msg("Error on client read")
			c.Connection.Close()
			s.removeClient(c)
			return
		}

		temp := make([]byte, cLen)
		temp = recBuff[:cLen]

		packet.Reset(handleClientData(temp, packet))
	}
}

func handleClientData(data []byte, incomingPacket *Packet) bool {
	packetLength := uint16(0)
	incomingPacket.SetBytes(data) // what was read from the buffer

	if incomingPacket.UnreadLength() >= 2 {
		packetLength = incomingPacket.ReadUint16()
		if packetLength <= 0 {
			return true
		}
	}

	for packetLength > 0 && packetLength <= incomingPacket.UnreadLength() {
		packetBytes := incomingPacket.ReadBytes(packetLength)
		newPacket := NewPacket(packetBytes)
		newPacket.client = incomingPacket.client

		// Handle data

		parseData(newPacket)

		packetLength = 0
		if incomingPacket.UnreadLength() >= 2 {
			packetLength = incomingPacket.UnreadLength()
			if packetLength <= 0 {
				return true
			}
		}
	}

	if packetLength <= 1 {
		return true
	}

	return false
}

func parseData(packet *Packet) {
	data := packet.GetRemainderBytes()
	// check if it is a command. At this stage we are assuming that the data exchange is in text
	if string(data[0]) == "/" {
		// this might be a directed chat
		payload := string(data[1:len(data)])
		private := strings.Index(payload, " ")
		// to send a message to a certain user, we need at least 2 params (destination user and a message)
		if private == -1 {
			// we actually got a space
			packet.client.Connection.Write([]byte("Invalid destination message\n"))
			return
		}

		textPayload := string(data[private+2 : len(data)]) // 1 position for / and another for space

		destID, err := uuid.Parse(payload[0:private])
		if err != nil {
			packet.client.Connection.Write([]byte("Invalid destination ID\n"))
			return
		}

		if destID == packet.client.ID {
			packet.client.Connection.Write([]byte("Cannot send message to yourself\n"))
			return
		}

		log.Info().
			Str("src", packet.client.ID.String()).
			Str("dst", destID.String()).
			Str("msg", textPayload).
			Msg("Destination client")

		if err := instance.BroadcastClient(destID, []byte(fmt.Sprintf("[Private from %s] %s\n", packet.client.ID.String(), textPayload))); err != nil {
			log.Err(err).
				Str("src", packet.client.ID.String()).
				Str("dst", destID.String()).
				Msg("Failed to send to user")
		}
	}
}

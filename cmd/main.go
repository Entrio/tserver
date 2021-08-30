package main

import (
	"fmt"
	"github.com/Entrio/tserver/internal"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"sync"
)

var (
	rwm     sync.Mutex
	clients map[uuid.UUID]*internal.Client
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	port := fmt.Sprintf("0.0.0.0:%d", 1337)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)

	if err != nil {
		log.Err(err).Msg("Failed to resolve address")
		return
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Err(err).Msg("Failed to create TCP listener")
		return
	}

	clients = make(map[uuid.UUID]*internal.Client)

	defer listener.Close()
	log.Info().Str("endpoint", port).Msg("Starting server")
	listenForConnections(listener)
}

func listenForConnections(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warn().Msg("Failed to accept tcp connection")
			log.Err(err).Msg("Error:")
			continue
		}
		newDescriptor(conn)
	}
}

func newDescriptor(connection net.Conn) {
	log.Info().
		Str("remote", connection.RemoteAddr().String()).
		Msg("Connection attempt")
	fmt.Println(fmt.Sprintf("New connection from %s", connection.RemoteAddr().String()))
	c := &internal.Client{
		ID:         uuid.New(),
		Connection: connection,
	}

	rwm.Lock()
	defer rwm.Unlock()
	log.Info().Str("UID", c.ID.String()).Msg("Connected to server")
	clients[c.ID] = c
}

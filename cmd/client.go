package main

import (
	"bufio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	args := os.Args

	if len(args) != 2 {
		log.Error().Msg("No address and port given, please use format host:port")
		return
	}

	con, err := net.Dial("tcp", args[1])
	if err != nil {
		log.Err(err).Msg("Failed to connect to remote host")
		return
	}
	defer con.Close()

	cReader := bufio.NewReader(os.Stdin)
	sReader := bufio.NewReader(con)

	go func() {
		cMessage, err := cReader.ReadString('\n')

		switch err {
		case nil:
			if _, err = con.Write([]byte(strings.TrimSpace(cMessage))); err != nil {
				log.Warn().Err(err).Msg("Failed to send message to the server")
			}
		case io.EOF:
			log.Warn().Msg("Connection was closed by the client")
			return
		default:
			log.Warn().Msg("Unknown error has occurred")
			return
		}
	}()

	for {
		serverResponse, err := sReader.ReadString('\n')
		switch err {
		case nil:
			log.Info().Str("msg", strings.TrimSpace(serverResponse)).Msg("Server:")
		case io.EOF:
			log.Warn().Msg("Server closed the remote connection")
			return
		default:
			log.Warn().Msg("Unknown server error")
			return
		}

	}
}

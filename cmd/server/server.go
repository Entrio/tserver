package main

import (
	"fmt"
	"github.com/Entrio/tserver/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	port := fmt.Sprintf("0.0.0.0:%d", 1337)

	if err := internal.NewServer(port).Start(); err != nil {
		log.Err(err).Msg("Failed to start server")
		return
	}
}

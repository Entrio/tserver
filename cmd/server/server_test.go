package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Entrio/tserver/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	port := fmt.Sprintf("0.0.0.0:%d", 1337)

	go func() {
		if err := internal.NewServer(port).Start(); err != nil {
			log.Err(err).Msg("Failed to start server")
			return
		}
	}()
	time.Sleep(time.Second * 1)
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: bufio.NewWriter(&buf)})
	f()
	return buf.String()
}

func TestServerRunning(t *testing.T) {
	servers := []struct {
		host string
		port int
	}{
		{"127.0.0.1", 1337},
	}

	for _, s := range servers {
		output := captureOutput(func() {
			t.Log("Capturing output")
			conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", s.host, s.port))
			if err != nil {
				t.Error("could not connect to server: ", err)
			}
			defer conn.Close()
		})
		t.Log(output)
		assert.Contains(t, output, "Welcome to FIRE PHOENIX testing server", "Check for server start message")
	}
}

func TestClientConnection(t *testing.T) {

}

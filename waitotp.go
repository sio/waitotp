package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pquerna/otp/totp"
	"golang.org/x/time/rate"
)

const (
	MaxMessageLength  = 10
	ListenAddr        = "127.0.0.1"
	ListenPort        = 19991
	TotpSecret        = "sampletotpsecret"
	ConnectionTimeout = 10 * time.Second
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ListenAddr, ListenPort))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	limit := rate.NewLimiter(1/2, 2)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handle(conn, limit)
	}
}

// Handle a single TCP session
func handle(conn net.Conn, limit *rate.Limiter) {
	// TCP hygiene
	conn.SetDeadline(time.Now().Add(ConnectionTimeout))
	defer conn.Close()

	// We only care about the first buffer
	buf := make([]byte, MaxMessageLength)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error: %s", err)
		return
	}
	code, err := parseTotpCode(buf[:n])
	if err != nil {
		log.Printf("invalid input from %s (%d bytes)", conn.RemoteAddr(), n)
		return
	}
	if !limit.Allow() {
		log.Printf("rate limit for TOTP verification exceeded")
		return
	}
	log.Printf("received TOTP code from %s (%d bytes)", conn.RemoteAddr(), n)
	if totp.Validate(code, TotpSecret) {
		log.Printf("TOTP code is valid, exiting...")
		os.Exit(0)
	}
}

// Drop any whitespace from input and ensure it contains a valid integer
func parseTotpCode(buf []byte) (string, error) {
	raw := string(buf)
	var builder strings.Builder
	builder.Grow(len(raw))
	for _, char := range raw {
		if !unicode.IsSpace(char) {
			builder.WriteRune(char)
		}
	}
	out := builder.String()
	_, err := strconv.Atoi(out)
	return out, err
}

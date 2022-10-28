package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/pquerna/otp/totp"
	"golang.org/x/time/rate"
)

const (
	MaxMessageLength = 10
	ListenAddr       = "127.0.0.1"
	ListenPort       = 19991
	TotpSecret       = "sampletotpsecret"
)

func main() {
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", ListenAddr, ListenPort))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	rateLimit := rate.NewLimiter(0.5, 2)
	for {
		buf := make([]byte, MaxMessageLength)
		n, addr, err := conn.ReadFrom(buf)
		if n == 0 || err != nil {
			continue
		}
		code, err := parseTotpCode(buf[:n])
		if err != nil {
			log.Printf("invalid input from %s (%d bytes)", addr, n)
			continue
		}
		if !rateLimit.Allow() {
			continue
		}
		log.Printf("received TOTP code from %s (%d bytes)", addr, n)
		if totp.Validate(code, TotpSecret) {
			log.Printf("TOTP code is valid, exiting...")
			os.Exit(0)
		}
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

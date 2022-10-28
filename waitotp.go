package main

import (
	"flag"
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
	ConnectionTimeout = 10 * time.Second
	TotpEnvVariable   = "TOTP_SECRET"
)

// Parse CLI arguments
type Config struct {
	Protocol   string
	IP         string
	Port       int
	TotpSecret string
}

func (conf *Config) Address() string {
	return fmt.Sprintf("%s:%d", conf.IP, conf.Port)
}
func (conf *Config) Parse() {
	conf.Protocol = "tcp" // TODO: other protocols are not supported yet
	flag.StringVar(&conf.IP, "ip", "127.0.0.1", "IP address to bind to")
	flag.IntVar(&conf.Port, "port", 20002, fmt.Sprintf("%s port to listen on", conf.Protocol))
	flag.Parse()
	var ok bool
	conf.TotpSecret, ok = os.LookupEnv(TotpEnvVariable)
	if !ok {
		log.Fatal("environment variable not defined: ", TotpEnvVariable)
	}
}

var conf Config // global variable

// CLI entrypoint
func main() {
	conf.Parse()

	listener, err := net.Listen(conf.Protocol, conf.Address())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("%s listening on %s", os.Args[0], conf.Address())

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
	if totp.Validate(code, conf.TotpSecret) {
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

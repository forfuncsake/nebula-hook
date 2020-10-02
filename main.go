package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
)

// A version string that can be set with
//
//     -ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string

const msgType = 99

func main() {
	configPath := flag.String("config", "", "Path to either a file or directory to load configuration from")
	configTest := flag.Bool("test", false, "Test the config and print the end result. Non zero exit indicates a faulty config")
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	server := flag.Bool("serve", false, "Run the nebula-hook server")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Println("-config flag must be set")
		flag.Usage()
		os.Exit(1)
	}

	config := nebula.NewConfig()
	err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("failed to load config: %s", err)
		os.Exit(1)
	}

	l := logrus.New()
	l.Out = os.Stdout
	c, err := nebula.Main(config, *configTest, Build, l, nil)

	switch v := err.(type) {
	case nebula.ContextualError:
		v.Log(l)
		os.Exit(1)
	case error:
		l.WithError(err).Error("Failed to start")
		os.Exit(1)
	}

	if *server {
		c.Hook(msgType, serve(c.Send))
	} else {
		dest := config.GetString("hook.server", "")
		if dest == "" {
			l.Error("hook.server is not set in config")
			os.Exit(1)
		}
		host, portString, err := net.SplitHostPort(dest)
		if err != nil {
			l.WithError(err).Error("Invalid hook.server address")
			os.Exit(1)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			l.WithError(err).Error("Invalid port in hook.server address")
		}

		c.Hook(msgType, handleResponse(l))
		go send(l, net.ParseIP(host), uint16(port), c.Send)
	}

	if !*configTest {
		c.Start()
		c.ShutdownBlock()
	}

	os.Exit(0)
}

func send(l *logrus.Logger, dest net.IP, port uint16, fn func(ip uint32, port uint16, t nebula.NebulaMessageSubType, payload []byte)) {
	ip := ip2int(dest)
	for i := 0; i < 1000; i++ {
		time.Sleep(5 * time.Second)

		l.WithField("value", i).Println("Hook payload sent")
		fn(ip, port, msgType, []byte(strconv.Itoa(i)))
	}
}

func handleResponse(l *logrus.Logger) func(b []byte) error {
	return func(b []byte) error {
		num, err := packetNumber(b)
		if err != nil {
			return err
		}

		l.WithField("value", num).Println("Hook payload received")
		return nil
	}
}

func serve(fn func(ip uint32, port uint16, t nebula.NebulaMessageSubType, payload []byte)) func(b []byte) error {
	return func(b []byte) error {
		num, err := packetNumber(b)
		if err != nil {
			return err
		}

		clientIP := ip2int(b[12:16])
		port := uint16(b[20])<<8 + uint16(b[21])

		num++
		fn(clientIP, uint16(port), msgType, []byte(strconv.Itoa(num)))

		return nil
	}
}

func packetNumber(b []byte) (int, error) {
	if len(b) < 24 {
		return 0, fmt.Errorf("expect more than 24 bytes, got %d", len(b))
	}

	// discard nebula firewall packet header
	b = b[24:]

	num, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse: %w", err)
	}

	return int(num), nil
}

func ip2int(ip []byte) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

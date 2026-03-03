package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	protocol := flag.String("proto", "tcp", "protocol (tcp or udp)")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	switch *protocol {
	case "tcp":
		runTCP(addr)
	case "udp":
		runUDP(addr)
	default:
		log.Fatalf("Unknown protocol: %s", *protocol)
	}
}

func runTCP(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP Echo Server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleTCPConn(conn)
	}
}

func handleTCPConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Connection closed: %s", conn.RemoteAddr())
			return
		}

		data := buf[:n]
		log.Printf("Received (%d bytes):\n%s", n, hex.Dump(data))
		conn.Write(data)
		log.Printf("Sent (%d bytes):\n%s", n, hex.Dump(data))
	}
}

func runUDP(addr string) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	log.Printf("UDP Echo Server listening on %s", addr)

	buf := make([]byte, 4096)
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		data := buf[:n]
		log.Printf("Received from %s (%d bytes):\n%s", remoteAddr, n, hex.Dump(data))
		conn.WriteTo(data, remoteAddr)
		log.Printf("Sent to %s (%d bytes):\n%s", remoteAddr, n, hex.Dump(data))
	}
}

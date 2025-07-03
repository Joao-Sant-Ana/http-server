package main

import (
	"log"
	"net"
	"time"
)

func main() {
	log.Println("Hello world")
	ln, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error: %v\n", err)
			break
		}

		handleConn(conn)
	}

}

func handleConn(c net.Conn) {
	defer c.Close()
	r := make([]byte, 256)
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := c.Read(r)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	
	response := "HTTP/1.1 200 OK\r\n\r\n"
	_, err = c.Write([]byte(response))
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
}
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, os.Interrupt)

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

		go handleConn(conn)
	}

	<-c
}

func handleConn(c net.Conn) {
	defer c.Close()
	r := make([]byte, 256)
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := c.Read(r)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	response := "HTTP/1.1 404 NOT FOUND\r\n\r\n"
	_, err = c.Write([]byte(response))
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
}

func getResponseStatus(s int) string {
	responseHeader := "HTTP/1.1 "

	switch s {
	// Informational
	case 100:
		responseHeader += "100 CONTINUE\r\n"
	case 101:
		responseHeader += "101 SWITCHING PROTOCOLS\r\n"
	case 102:
		responseHeader += "102 PROCESSING\r\n"

	// Successful
	case 200:
		responseHeader += "200 OK\r\n"
	case 201:
		responseHeader += "201 CREATED\r\n"
	case 202:
		responseHeader += "202 ACCEPTED\r\n"
	case 203:
		responseHeader += "203 NON-AUTHORITATIVE INFORMATION\r\n"
	case 204:
		responseHeader += "204 NO CONTENT\r\n"
	case 205:
		responseHeader += "205 RESET CONTENT\r\n"
	case 206:
		responseHeader += "206 PARTIAL CONTENT\r\n"
	case 207:
		responseHeader += "207 MULTI-STATUS\r\n"
	case 208:
		responseHeader += "208 ALREADY REPORTED\r\n"

	// Redirection
	case 300:
		responseHeader += "300 MULTIPLE CHOICES\r\n"
	case 301:
		responseHeader += "301 MOVED PERMANENTLY\r\n"
	case 302:
		responseHeader += "302 FOUND\r\n"
	case 303:
		responseHeader += "303 SEE OTHER\r\n"
	case 304:
		responseHeader += "304 NOT MODIFIED\r\n"
	case 305:
		responseHeader += "305 USE PROXY\r\n"
	case 307:
		responseHeader += "307 TEMPORARY REDIRECT\r\n"
	case 308:
		responseHeader += "308 PERMANENT REDIRECT\r\n"

	// Client Error
	case 400:
		responseHeader += "400 BAD REQUEST\r\n"
	case 401:
		responseHeader += "401 UNAUTHORIZED\r\n"
	case 402:
		responseHeader += "402 PAYMENT REQUIRED\r\n"
	case 403:
		responseHeader += "403 FORBIDDEN\r\n"
	case 404:
		responseHeader += "404 NOT FOUND\r\n"
	case 405:
		responseHeader += "405 METHOD NOT ALLOWED\r\n"
	case 406:
		responseHeader += "406 NOT ACCEPTABLE\r\n"
	case 407:
		responseHeader += "407 PROXY AUTHENTICATION REQUIRED\r\n"
	case 408:
		responseHeader += "408 REQUEST TIMEOUT\r\n"
	case 409:
		responseHeader += "409 CONFLICT\r\n"

	// Server Error
	case 500:
		responseHeader += "500 INTERNAL SERVER ERROR\r\n"
	case 501:
		responseHeader += "501 NOT IMPLEMENTED\r\n"
	case 502:
		responseHeader += "502 BAD GATEWAY\r\n"
	case 503:
		responseHeader += "503 SERVICE UNAVAILABLE\r\n"
	case 504:
		responseHeader += "504 GATEWAY TIMEOUT\r\n"
	case 505:
		responseHeader += "505 HTTP VERSION NOT SUPPORTED\r\n"

	default:
		responseHeader += "500 INTERNAL SERVER ERROR\r\n"
	}

	return responseHeader
}

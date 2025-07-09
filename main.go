package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Path string `yaml:"path"`
	Address string `yaml:"address"`
}

type Config struct {
	Servers map[string] Server `yaml:"servers"`
}

var requests = make(chan net.Conn, 512)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, os.Interrupt)

	_, err := os.ReadDir("./sites-available")
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir("./sites-available", 0755)
			if err != nil {
				log.Fatalf("Error creating sites-available folder: %v", err)
			}
			file, err := os.Create("./sites-available/default.yaml")
			if err != nil {
				log.Fatalf("Error creating default file: %v", err)
			}
			defaultConfig := &Config{
				Servers: map[string]Server{
					"default": {
						Path: "localhost:3000",
						Address: "",
					},
				},
			}
			marshaledConfig, err := yaml.Marshal(defaultConfig)
			if err != nil {
				log.Fatalf("Error formating config: %v", err)
			}

			if _, err := file.Write(marshaledConfig); err != nil {
				log.Fatalf("Error writing default config file: %v", err)
			}
		} else {
			log.Fatalf("Error while reading the sites-available dir: %v", err)
		}
	}

	files, err := os.ReadDir("./sites-enabled")
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir("./sites-enabled", 0755); err != nil {
				log.Fatalf("Error creating sites-enabled folder: %v", err)
			}

			if err := os.Link("./sites-available/default.yaml", "./sites-enabled/default.yaml"); err != nil {
				log.Fatalf("Error linking default files: %v", err)
			}
		} else {
			log.Fatalf("Error while reading the sites-enabled dir: %v", err)
		}
	}

	var addresses []Config

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		config, err := loadConfig(fmt.Sprintf("./sites-enabled/%s", file.Name()))
		if err != nil {
			log.Fatalf("Error while getting config: %v", err)
		}
		log.Println(config)

		addresses = append(addresses, config)
	}

	ln, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	go reverseProxy(addresses)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("Error: %v\n", err)
				break
			}
			
			requests <- conn
		}
	}()
	
	<-c
}

func reverseProxy(addresses []Config) {
	for conn := range requests {
		go func() {
			defer conn.Close()

			reader := bufio.NewReader(conn)

			requestLine, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Failed to read request line: %v", err)
				return
			}

			var headers []string
			var host string

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Failed to read headers: %v", err)
					return
				}

				line = strings.TrimSpace(line)

				if line == "" {
					break
				}

				headers = append(headers, line)

				if strings.HasPrefix(strings.ToLower(line), "host:") {
					host = strings.TrimSpace(strings.TrimPrefix(line, "Host:"))
					host = strings.TrimSpace(strings.TrimPrefix(host, "host:"))
				}
			}

			var targetServer *Server
			for _, address := range addresses {
				for _, value := range address.Servers {
					if value.Address == host {
						targetServer = &value
						break
					}
				}
				if targetServer != nil {
					break
				}
			}

			if targetServer == nil {
				log.Printf("No server found for host: %s", host)
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"))
				return
			}

			targetConn, err := net.Dial("tcp", targetServer.Path)
			if err != nil {
				log.Printf("Error connecting to target server: %v", err)
				conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n"))
				return
			}
			defer targetConn.Close()

			_, err = targetConn.Write([]byte(requestLine))
			if err != nil {
				log.Printf("Error writing request line: %v", err)
				return
			}

			for _, header := range headers {
				_, err = targetConn.Write([]byte(header + "\r\n"))
				if err != nil {
					log.Printf("Error writing header: %v", err)
					return
				}
			}

			_, err = targetConn.Write([]byte("\r\n"))
			if err != nil {
				log.Printf("Error writing header terminator: %v", err)
				return
			}

			go func() {
				defer targetConn.Close()
				io.Copy(targetConn, reader)
			}()

			io.Copy(conn, targetConn)
		}()
	}
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

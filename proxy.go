/*

Proxy Socks5 Golang Websocket By Farell Aditya
Date Thu Feb 13 06:36:21 UTC 2025
Only For Linux Usage

*/

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// Application version
const Version = "1.0.0"

// Config structure
type Config struct {
	SOCKS5Port string
	SSHPort    string
}

// Function to authenticate SSH credentials
func authenticateSSH(username, password, sshPort string) bool {
	sshAddress := fmt.Sprintf("127.0.0.1:%s", sshPort)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", sshAddress, config)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// Custom SSH Authentication struct
type sshAuth struct {
	SSHPort string
}

// Implement proxy.Authenticator for SSH authentication
func (s *sshAuth) Authenticate(username, password string) error {
	if authenticateSSH(username, password, s.SSHPort) {
		return nil // Authentication successful
	}
	return fmt.Errorf("authentication failed")
}

func main() {
	// Command-line flags
	socks5Port := flag.String("p", "1080", "Port for SOCKS5 Proxy (default: 1080)")
	sshPort := flag.String("f", "22", "Port for SSH Server (default: 22)")
	version := flag.Bool("version", false, "Show application version")
	help := flag.Bool("h", false, "Show help menu")

	flag.Parse()

	// Handle --version
	if *version {
		fmt.Printf("SOCKS5 Proxy with SSH Auth - Version %s\n", Version)
		os.Exit(0)
	}

	// Handle --help or -h
	if *help {
		fmt.Println("Usage:")
		fmt.Println("  socks5_proxy [OPTIONS]")
		fmt.Println("\nOptions:")
		fmt.Println("  -p <port>      Port for SOCKS5 Proxy (default: 1080)")
		fmt.Println("  -f <port>      Port for SSH Server (default: 22)")
		fmt.Println("  --version      Show application version")
		fmt.Println("  -h, --help     Show this help menu")
		os.Exit(0)
	}

	config := Config{
		SOCKS5Port: *socks5Port,
		SSHPort:    *sshPort,
	}

	// Start SOCKS5 proxy server
	listener, err := net.Listen("tcp", ":"+config.SOCKS5Port)
	if err != nil {
		log.Fatal("Failed to start SOCKS5 server:", err)
	}
	defer listener.Close()

	log.Printf("SOCKS5 proxy server running on port %s (SSH auth on port %s)\n", config.SOCKS5Port, config.SSHPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleClient(conn, config)
	}
}

// Function to handle SOCKS5 authentication and proxying
func handleClient(client net.Conn, config Config) {
	defer client.Close()

	// Setup authentication
	auth := &proxy.Auth{}
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:"+config.SOCKS5Port, auth, proxy.Direct)
	if err != nil {
		log.Println("Failed to create SOCKS5 dialer:", err)
		return
	}

	target, err := dialer.Dial("tcp", client.RemoteAddr().String())
	if err != nil {
		log.Println("Failed to connect to target:", err)
		return
	}
	defer target.Close()

	// Relay data between client and target
	go io.Copy(target, client)
	io.Copy(client, target)
}

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type client struct {
	conn   net.Conn
	name   string
	server *server
}

type server struct {
	listenAddr string
	ln         net.Listener
	clients    map[*client]struct{}
	quitch     chan struct{}
	logger     *log.Logger
}

func NewServer(listenAddr string, logger *log.Logger) *server {
	return &server{
		listenAddr: listenAddr,
		clients:    make(map[*client]struct{}),
		quitch:     make(chan struct{}),
		logger:     logger,
	}
}

func (s *server) start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		s.logger.Printf("Error starting server: %v", err)
		return err
	}
	s.ln = ln

	fmt.Println("Lestning on the port:", s.listenAddr)
	s.logger.Println("Listening on the port:", s.listenAddr)
	s.logger.Printf("Server started and listening on %s", s.listenAddr)

	go s.acceptloop()

	<-s.quitch

	return nil
}

func (s *server) acceptloop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			s.logger.Printf("Accept error: %v", err)
			continue
		}
		c := &client{
			conn:   conn,
			server: s,
		}
		s.clients[c] = struct{}{}
		go c.handle()
	}
}

func (c *client) handle() {
	defer func() {
		c.conn.Close()
		c.server.removeClient(c)
	}()

	// Send welcome message
	welcomeMessage, err := os.ReadFile("welcome.txt")
	if err != nil {
		c.server.logger.Printf("Failed to read welcome message: %v", err)
		return
	}
	_, err = c.conn.Write(welcomeMessage)
	if err != nil {
		c.server.logger.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Read client's name
	nameBuf := make([]byte, 1024)
	n, err := c.conn.Read(nameBuf)
	if err != nil {
		c.server.logger.Printf("Read error: %v", err)
		return
	}
	c.name = strings.TrimSpace(string(nameBuf[:n]))

	// Inform all clients that a new client joined
	joinMsg := fmt.Sprintf("[%s] %s has joined the chat...\n", time.Now().Format("2006-01-02 15:04:05"), c.name)
	c.server.broadcastMessage(joinMsg)
	c.server.logger.Print(joinMsg)

	// Read messages from client
	msgBuf := make([]byte, 1024)
	for {
		n, err := c.conn.Read(msgBuf)
		if err != nil {
			if err.Error() == "EOF" {
				c.server.logger.Printf("Client %s disconnected", c.name)
			} else {
				c.server.logger.Printf("Read error from %s: %v", c.name, err)
			}
			break
		}
		msg := strings.TrimSpace(string(msgBuf[:n]))
		if msg != "" {
			broadcastMsg := fmt.Sprintf("[%s] [%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), c.name, msg)
			c.server.broadcastMessage(broadcastMsg)
			c.server.logger.Print(broadcastMsg)
		}
	}
}

func (s *server) removeClient(c *client) {
	delete(s.clients, c)
	leaveMsg := fmt.Sprintf("[%s] %s has left the chat...\n", time.Now().Format("2006-01-02 15:04:05"), c.name)
	s.broadcastMessage(leaveMsg)
	s.logger.Print(leaveMsg)
}

func (s *server) broadcastMessage(msg string) {
	for client := range s.clients {
		_, err := client.conn.Write([]byte(msg))
		if err != nil {
			s.logger.Printf("Failed to send message to %s: %v", client.name, err)
		}
	}
}

func main() {
	if len(os.Args) > 3 || (len(os.Args) == 3 && os.Args[2] != "localhost") {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := ":8989"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	// Open log file
	logFile, err := os.OpenFile("chatlog.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Create logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Add a test log entry to ensure logger is working
	logger.Println("Server is starting up...")

	server := NewServer(port, logger)
	logger.Fatal(server.start())
}

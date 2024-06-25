package main

import (
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestServerStart(t *testing.T) {
	// Create a log file for the test
	logFile, err := os.OpenFile("chatlog_test.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Initialize and start server
	server := NewServer(":9999", logger)
	go func() {
		if err := server.start(); err != nil {
			t.Fatalf("Error starting server: %v", err)
		}
	}()

	// Give some time for the server to start
	time.Sleep(time.Second)

	// Try connecting to the server
	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Errorf("Error connecting to server: %v", err)
		return
	}
	defer conn.Close()
	logger.Println("Successfully connected to the server")

	// Check the log file for confirmation
	content, err := os.ReadFile("chatlog_test.txt")
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "Successfully connected to the server") {
		t.Error("Expected log message not found in log file")
	}
}

func TestClientConnection(t *testing.T) {
	// Create a log file for the test
	logFile, err := os.OpenFile("chatlog_test.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Initialize and start server
	server := NewServer(":9998", logger)
	go func() {
		if err := server.start(); err != nil {
			t.Fatalf("Error starting server: %v", err)
		}
	}()

	// Give some time for the server to start
	time.Sleep(time.Second)

	// Create a client and try connecting to the server
	conn, err := net.Dial("tcp", "localhost:9998")
	if err != nil {
		t.Errorf("Error connecting to server: %v", err)
		return
	}
	defer conn.Close()
	logger.Println("Successfully connected to the server as a client")

	// Check the log file for confirmation
	content, err := os.ReadFile("chatlog_test.txt")
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "Successfully connected to the server as a client") {
		t.Error("Expected log message not found in log file")
	}
}

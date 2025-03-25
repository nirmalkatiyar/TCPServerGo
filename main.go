package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
)

// Do handles incoming connections, reads data, processes it, and sends a response.
func Do(conn net.Conn) { // Handle incoming connection
	defer conn.Close() // Ensure the connection is closed when done

	// Set a read deadline to prevent hanging
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println("Failed to read from the connection:", err)
		return
	}

	// Simulate processing time
	time.Sleep(1 * time.Second)
	// Write response
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello, from [TCP Server]"))
}

func main() {
	// Log server start
	fmt.Println("Server starting..")
	
	// Set backlog size to 10
	tcpServer, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer tcpServer.Close() // Ensure the server is closed when done

	// Channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\nShutting down the server...") // Log server shutdown
		tcpServer.Close()
		os.Exit(0)
	}()

	// Define a semaphore with a maximum of 10 concurrent connections
	maxConnections := 10
	semaphore := make(chan struct{}, maxConnections)

	// Create a pool of worker goroutines
	for {
		con, err := tcpServer.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue // Skip to the next iteration on error
		}

		// Acquire a slot in the semaphore
		semaphore <- struct{}{} // Block if the limit is reached

		go func(conn net.Conn) {
			defer func() { <-semaphore }() // Release the slot when done
			Do(conn) // Handle the connection concurrently
		}(con) // Pass the connection to the goroutine
	}
}

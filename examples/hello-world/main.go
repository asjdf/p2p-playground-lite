package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	hostname, _ := os.Hostname()
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	fmt.Printf("Hello from P2P Playground!\n")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Hostname: %s\n", hostname)
	fmt.Printf("Started at: %s\n\n", time.Now().Format(time.RFC3339))

	// Keep running and print heartbeat
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("[%s] Application is running...\n", t.Format("15:04:05"))
		}
	}
}

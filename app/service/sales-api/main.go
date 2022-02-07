package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	l := log.New(os.Stdout, "TEST", log.Lshortfile|log.Ldate|log.Ltime)

	l.Println("Starting Service")
	defer l.Println("Stopped Service")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
}

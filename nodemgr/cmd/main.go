package main

import (
	"log"
	"nodemgr/internal/core/domain"
)

func main() {
	log.Printf("Starting Node Manager...")

	n := domain.NewNode("example-node")
	_ = n
}

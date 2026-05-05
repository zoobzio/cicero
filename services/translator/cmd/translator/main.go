// Package main is the entry point for the translator gRPC service.
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/zoobz-io/cicero/proto/translator"
	"github.com/zoobz-io/cicero/translator"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := getEnv("TRANSLATOR_LISTEN_ADDR", ":9091")
	libreTranslateAddr := getEnv("LIBRETRANSLATE_ADDR", "http://localhost:5000")

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	srv := translator.NewServer(libreTranslateAddr)
	grpcServer := grpc.NewServer()
	pb.RegisterTranslatorServiceServer(grpcServer, srv)

	log.Printf("translator listening on %s (libretranslate: %s)", addr, libreTranslateAddr)
	return grpcServer.Serve(listener)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

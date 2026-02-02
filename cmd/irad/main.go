package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cchirag/ira/internal/config"
	"github.com/cchirag/ira/internal/services/root"
	"github.com/cchirag/ira/internal/services/session"
	rootProtoV1 "github.com/cchirag/ira/proto/gen/services/v1/root"
	sessionProtov1 "github.com/cchirag/ira/proto/gen/services/v1/session"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logFile, err := os.OpenFile(config.Current.DaemonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Starting daemon in %s mode", config.Current.Mode)
	log.Printf("Socket: %s", config.Current.DaemonSocketPath)
	log.Printf("Database: %s", config.Current.DaemonDBPath)
	log.Printf("Log: %s", config.Current.DaemonLogPath)

	db, err := bbolt.Open(config.Current.DaemonDBPath, 0600, nil)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	log.Println("Database opened successfully")

	socketPath := config.Current.DaemonSocketPath

	if runtime.GOOS != "windows" {
		if _, err = os.Stat(socketPath); err == nil {
			if err = os.Remove(socketPath); err != nil {
				log.Fatalf("failed to remove old socket: %v", err)
			}
			log.Println("Removed existing socket file")
		}
	}

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("failed to listen on socket: %v", err)
	}
	defer lis.Close()

	if runtime.GOOS != "windows" {
		if err := os.Chmod(socketPath, 0600); err != nil {
			log.Printf("warning: failed to set socket permissions: %v", err)
		}
	}

	log.Printf("Listening on %s", socketPath)

	closeCh := make(chan struct{})

	grpcServer := grpc.NewServer()

	rootProtoV1.RegisterRootServiceServer(grpcServer, &root.Service{
		Db:      db,
		CloseCh: closeCh,
	})

	sessionProtov1.RegisterSessionServiceServer(grpcServer, &session.Service{
		Db: db,
	})

	reflection.Register(grpcServer)
	log.Println("gRPC services registered")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-closeCh:
			log.Println("Received shutdown signal from close channel")
			grpcServer.GracefulStop()
		case sig := <-sigCh:
			log.Printf("Received signal: %v", sig)
			grpcServer.GracefulStop()
		}

		if runtime.GOOS != "windows" {
			os.Remove(socketPath)
			log.Println("Cleaned up socket file")
		}
	}()

	log.Println("Daemon started successfully")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

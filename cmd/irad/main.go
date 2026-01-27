package main

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/cchirag/ira/internal/services/root"
	protov1 "github.com/cchirag/ira/proto/gen/services/v1"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const PORT = ":50051"

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	appConfigPath := filepath.Join(configDir, "ira")
	db, err := bbolt.Open(appConfigPath, 0600, nil)
	if err != nil {
		log.Fatalf("error opening the db: %s", err.Error())
	}
	defer db.Close()

	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	protov1.RegisterRootServiceServer(grpcServer, &root.Service{
		Db: db,
	})
	reflection.Register(grpcServer)

	log.Printf("🚀 IRA gRPC server listening on port %s", PORT)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err.Error())
	}
}

package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	"github.com/shanehowearth/entain-master/sport/db"
	"github.com/shanehowearth/entain-master/sport/proto/sporting"
	"github.com/shanehowearth/entain-master/sport/service"
	"google.golang.org/grpc"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", "localhost:9005", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":9005")
	if err != nil {
		return err
	}

	sportingDB, err := sql.Open("sqlite3", "./db/events.db")
	if err != nil {
		return err
	}

	eventsRepo := db.NewEventsRepo(sportingDB)
	if err := eventsRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	sporting.RegisterSportingServer(
		grpcServer,
		service.NewSportingService(
			eventsRepo,
		),
	)

	log.Printf("gRPC server listening on: %s\n", *grpcEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}

package main

import (
	"log"
	"net"

	"forge/project-service/config"
	"forge/project-service/database"
	"forge/project-service/handler"
	"forge/project-service/repository"
	projectservice "forge/project-service/service"
	projectpb "forge/proto"
		_ "net/http/pprof"
	"net/http"

	_ "go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("database connection failed:", err)
	}
	defer db.Close()

	if err := database.CreateTables(db); err != nil {
		log.Fatal("table initialization failed:", err)
	}

	projectRepository := repository.NewProjectRepository(db)

	service := projectservice.NewProjectService(
		projectRepository,
		cfg.PublicURL,
	)

	projectHandler := handler.NewProjectHandler(service)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("listener creation failed:", err)
	}

	go func() {
	log.Println("pprof running on :6061")
	if err := http.ListenAndServe("localhost:6061", nil); err != nil {
		log.Printf("pprof server stopped: %v", err)
	}
    }()

	grpcServer := grpc.NewServer()

	projectpb.RegisterProjectServiceServer(
		grpcServer,
		projectHandler,
	)

	// grpcurl testing ke liye
	reflection.Register(grpcServer)

	log.Printf(
		"Project Service listening on gRPC port %s",
		cfg.GRPCPort,
	)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("gRPC server failed:", err)
	}
}
package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/longtk26/simple_bank/api"
	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/gapi"
	"github.com/longtk26/simple_bank/pb"
	"github.com/longtk26/simple_bank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create grpc server:", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterSimpleBankServer(grpcServer, server)

	// Register reflection service on gRPC server.
	// Reflection is a feature that allows clients to query available RPCs
	reflection.Register(grpcServer)

	listner, err := net.Listen("tcp", config.GRPCDomain)
	if err != nil {
		log.Fatal("cannot start grpc server:", err)
	}
	
	log.Println("starting grpc server on ", config.GRPCDomain)
	err = grpcServer.Serve(listner)

	if err != nil {
		log.Fatal("cannot start grpc server:", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create grpc server:", err)
	}

	jsonOptions := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOptions)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal("cannot register gateway server:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listner, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start grpc gateway server:", err)
	}
	
	log.Println("starting http gateway server on ", listner.Addr().String())
	err = http.Serve(listner, mux)

	if err != nil {
		log.Fatal("cannot start http gateway server:", err)
	}
}
package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/longtk26/simple_bank/api"
	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/gapi"
	"github.com/longtk26/simple_bank/pb"
	"github.com/longtk26/simple_bank/util"
	"github.com/longtk26/simple_bank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	
	config, err := util.LoadConfig(".")
	
	if config.ENV == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if err != nil {
		log.Fatal().Msgf("cannot load config %s", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db %s", err)
	}
	defer conn.Close()

	runDBMigrations(config.MigrationUrl, config.DBSource)

	store := db.NewStore(conn)
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisURL,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runGatewayServer(config, store, taskDistributor)
	go runTaskProcessor(redisOpt, store)
	runGrpcServer(config, store, taskDistributor)
}

func runDBMigrations(migrationUrl string, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)

	if err != nil {
		
		log.Fatal().Msgf("cannot create migration %s", err)
	}

	err = migration.Up()

	if err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("cannot run migration up %s", err)
	}

	log.Info().Msg("DB migration up successful")
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal().Msgf("cannot create server %s", err)
	}

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal().Msgf("cannot start server %s", err)
	}
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msgf("cannot create grpc server %s", err)
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcServer, server)

	// Register reflection service on gRPC server.
	// Reflection is a feature that allows clients to query available RPCs
	reflection.Register(grpcServer)

	listner, err := net.Listen("tcp", config.GRPCDomain)
	if err != nil {
		log.Fatal().Msgf("cannot start grpc server %s", err)
	}
	
	log.Info().Msgf("starting grpc server on %s", listner.Addr().String())
	err = grpcServer.Serve(listner)

	if err != nil {
		log.Fatal().Msgf("cannot start grpc server %s", err)
	}
}

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msgf("cannot create grpc server %s", err)
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
		log.Fatal().Msgf("cannot register gateway server %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listner, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot start grpc gateway server %s", err)
	}
	
	log.Info().Msgf("starting http gateway server on %s", listner.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listner, handler)

	if err != nil {
		log.Fatal().Msgf("cannot start http gateway server %s", err)
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().Msg("=== Starting task processor ===")
	err := taskProcessor.Start()	

	if err != nil {
		log.Fatal().Msgf("cannot start task processor %s", err)
	}
}
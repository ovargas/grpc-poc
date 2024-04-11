package main

import (
	"context"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	dummyv1 "grpc-poc/api/dummy/v1"
	"grpc-poc/cmd/dummy"
	"grpc-poc/cmd/system"
	"log"
	"log/slog"
	"net"
	"net/http"
)

func main() {
	lis, err := net.Listen("tcp", ":8088")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	v := []system.OrganizationVerification{
		system.OrganizationRequired,
		ValidOrganizations("acme", "foo", "bar"),
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_logging.UnaryServerInterceptor(LoggerWrapper()),
			grpc_recovery.UnaryServerInterceptor(),
			system.OrganizationUnaryServerInterceptor(v...),
		),
		grpc.ChainStreamInterceptor(
			grpc_logging.StreamServerInterceptor(LoggerWrapper()),
			grpc_recovery.StreamServerInterceptor(),
			system.OrganizationStreamServerInterceptor(v...)),
	)

	dummyv1.RegisterDummyServiceServer(s, dummy.New())

	go func() {
		log.Println("Serving gRPC on 0.0.0.0:8088")
		log.Fatal(s.Serve(lis))
	}()

	c, err := grpc.NewClient("0.0.0.0:8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(system.WithIncomingOrganizationHeader))

	if err := dummyv1.RegisterDummyServiceHandler(context.Background(), mux, c); err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: system.OrganizationMiddleware(mux),
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8080")
	log.Fatalln(server.ListenAndServe())
}

func ValidOrganizations(orgs ...string) func(context.Context, string) error {
	return func(ctx context.Context, organization string) error {
		for _, org := range orgs {
			if org == organization {
				return nil
			}
		}
		return status.Error(codes.PermissionDenied, "organization not allowed")
	}
}

type logger struct{}

func LoggerWrapper() grpc_logging.Logger {
	return &logger{}
}

func (l *logger) Log(ctx context.Context, level grpc_logging.Level, msg string, attrs ...any) {
	slog.Log(ctx, slog.Level(level), msg, attrs...)
}

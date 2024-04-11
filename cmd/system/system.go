package system

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

type (
	organizationKey     struct{}
	streamServerWrapper struct {
		grpc.ServerStream
		ctx context.Context
	}
)

const OrganizationHeaderName = "X-Organization"

func (s *streamServerWrapper) Context() context.Context {
	return s.ctx
}

func WithOrganization(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, organizationKey{}, domain)
}

func WithIncomingOrganizationHeader(key string) (string, bool) {
	if key == OrganizationHeaderName {
		return key, true
	}
	return runtime.DefaultHeaderMatcher(key)
}

func Organization(ctx context.Context) string {
	v := ctx.Value(organizationKey{})
	if v == nil {
		panic("organization not found in context")
	}
	return v.(string)
}

func OrganizationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get(OrganizationHeaderName) != "" {
			h.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(r.URL.Path, "/api", 2)
		if len(parts) < 2 || parts[0] == "" {
			h.ServeHTTP(w, r)
			return
		}

		r.Header.Add(OrganizationHeaderName, strings.TrimPrefix(parts[0], "/"))

		http.StripPrefix(parts[0], h).ServeHTTP(w, r)
	})
}

func OrganizationStreamServerInterceptor(validators ...OrganizationVerification) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		o := getOrganizationFromHeaders(ss.Context())
		for _, v := range validators {
			if err := v(ss.Context(), o); err != nil {
				return err
			}
		}

		return handler(srv, &streamServerWrapper{
			ServerStream: ss,
			ctx:          WithOrganization(ss.Context(), o),
		})
	}
}

type OrganizationVerification func(ctx context.Context, organization string) error

func OrganizationUnaryServerInterceptor(validators ...OrganizationVerification) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		o := getOrganizationFromHeaders(ctx)
		for _, v := range validators {
			if err := v(ctx, o); err != nil {
				return nil, err
			}
		}

		ctx = WithOrganization(ctx, o)
		return handler(ctx, req)
	}
}

func OrganizationRequired(_ context.Context, organization string) error {
	if organization == "" {
		return status.Error(codes.InvalidArgument, "organization is required")
	}
	return nil
}

func getOrganizationFromHeaders(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	organizations := md.Get(OrganizationHeaderName)
	if len(organizations) != 0 {
		return organizations[0]
	}
	return ""
}

func LoggingUnaryServerInterceptor(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return handler(ctx, req)
}

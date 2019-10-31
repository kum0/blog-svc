package transport

import (
	"context"
	"github.com/Zhan9Yunhua/blog-svr/servers/usersvc/endpoints"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitGrpcTransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userPb "github.com/Zhan9Yunhua/blog-svr/pb/user"
	"github.com/Zhan9Yunhua/blog-svr/servers/usersvc/service"
	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin/zipkin-go"
	"google.golang.org/grpc"
)

type grpcServer struct {
	getUser kitGrpcTransport.Handler `json:""`
}

func (s *grpcServer) GetUser(ctx context.Context, req *userPb.GetUserRequest) (*userPb.GetUserReply, error) {
	_, rp, err := s.getUser.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcEncodeError(err)
	}
	rep := rp.(*userPb.GetUserReply)
	return &userPb.GetUserReply{Uid: rep.Uid}, nil
}

func NewGRPCClient(conn *grpc.ClientConn, otTracer opentracing.Tracer, zipkinTracer *zipkin.Tracer,
	logger log.Logger) service.IUserService {
	// limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
	// zipkinClient := kitZipkin.GRPCClientTrace(zipkinTracer)
	//
	// options := []kitGrpcTransport.ClientOption{
	// 	zipkinClient,
	// }

	var getUserEndpoint endpoint.Endpoint
	{
		getUserEndpoint = kitGrpcTransport.NewClient(
			conn,
			"pb.Usersvc",
			"GetUser",
			encodeGRPCGetUserRequest,
			decodeGRPCGetUserResponse,
			userPb.GetUserReply{},
			// append(options, kitGrpctransport.ClientBefore(kitOpentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		// getUserEndpoint = kitOpentracing.TraceClient(otTracer, "GetUser")(getUserEndpoint)
		// getUserEndpoint = limiter(getUserEndpoint)
	}

	return endpoints.Endponits{
		GetUserEP: getUserEndpoint,
	}
}

func MakeGRPCServer(endpoints endpoints.Endponits, otTracer opentracing.Tracer, zipkinTracer *zipkin.Tracer,
	logger log.Logger) userPb.UsersvcServer {
	// zipkinServer := kitZipkin.GRPCServerTrace(zipkinTracer)
	//
	// options := []kitGrpcTransport.ServerOption{
	// 	zipkinServer,
	// }

	return &grpcServer{
		getUser: kitGrpcTransport.NewServer(
			endpoints.GetUserEP,
			decodeGRPCGetUserRequest,
			encodeGRPCGetUserResponse,
			// append(options, kitGrpcTransport.ServerBefore(kitOpentracing.GRPCToContext(otTracer, "GetUser",
			// 	logger)))...,
		),
	}
}

func grpcEncodeError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		return status.Error(st.Code(), st.Message())
	}
	switch err {
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

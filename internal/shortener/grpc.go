package shortener

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/nessai1/linkshortener/proto"
)

// AddLink TODO
func (application *Application) AddLink(context.Context, *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddLink not implemented")
}

// GetLink TODO
func (application *Application) GetLink(context.Context, *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLink not implemented")
}

// AddLinkBatch TODO
func (application *Application) AddLinkBatch(context.Context, *pb.AddLinkBatchRequest) (*pb.AddLinkBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddLinkBatch not implemented")
}

// GetUserLinks TODO
func (application *Application) GetUserLinks(context.Context, *pb.GetUserLinksRequest) (*pb.GetUserLinksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserLinks not implemented")
}

// DeleteLink TODO
func (application *Application) DeleteLink(context.Context, *pb.DeleteLinkRequest) (*pb.DeleteLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLink not implemented")
}

// GetServiceStats TODO
func (application *Application) GetServiceStats(context.Context, *pb.GetServiceStatsRequest) (*pb.GetUserLinksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServiceStats not implemented")
}

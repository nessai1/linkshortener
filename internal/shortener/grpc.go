package shortener

import (
	"context"
	"errors"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener/encoder"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	pb "github.com/nessai1/linkshortener/proto"
)

func (application *Application) GetGRPCAddr() string {
	return application.config.GRPCAddr
}

func (application *Application) GetGRPCInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		getRegisterInterceptor([]string{
			"AddLink",
			"AddLinkBatch",
		}),

		getLoginInterceptor([]string{
			"GetUserLinks",
			"DeleteLink",
		}),
	}
}

func (application *Application) RegisterGRPCService(server *grpc.Server) error {
	pb.RegisterShortenerServiceServer(server, application)

	return nil
}

// AddLink add new link to service by gRPC
func (application *Application) AddLink(ctx context.Context, r *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	userUUIDCtxValue := ctx.Value(app.ContextUserUUIDKey)
	if userUUIDCtxValue == nil {
		application.logger.Error("[gRPC] No user UUID assigned")
		return nil, status.Error(codes.PermissionDenied, "No user UUID assigned")
	}

	userUUID := userUUIDCtxValue.(app.UserUUID)

	if !validateURL([]byte(r.Link)) {
		application.logger.Debug("[gRPC] Client sends invalid url", zap.String("url", r.Link))
		return nil, status.Error(codes.InvalidArgument, "Invalid link assigned")
	}

	hash, err := application.createResource(
		ctx,
		linkstorage.Link{
			Value:     r.Link,
			OwnerUUID: string(userUUID),
		},
	)

	if err != nil {
		if errors.Is(err, linkstorage.ErrURLIntersection) {
			application.logger.Debug(fmt.Sprintf("[gRPC] User insert duplicate url: %s", r.Link))
			return nil, status.Error(codes.AlreadyExists, "Got duplicate URL")
		} else {
			application.logger.Error(fmt.Sprintf("[gRPC] Cannot create resource for \"%s\"", r.Link), zap.Error(err))
			return nil, status.Error(codes.Internal, "Cannot create resource")
		}
	}

	hashedLink := application.buildTokenTail() + hash

	var response pb.AddLinkResponse
	response.Hash = hashedLink
	application.logger.Debug("[gRPC] Client success add URL", zap.String("URL", hashedLink))

	return &response, nil
}

// GetLink returns not deleted link if exists by his hash
func (application *Application) GetLink(ctx context.Context, r *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	if r.Hash == "" {
		application.logger.Debug("[gRPC] Client sends empty hash")
		return nil, status.Error(codes.InvalidArgument, "Client sends empty hash")
	}

	link, ok := application.storage.Get(ctx, r.Hash)
	if !ok {
		application.logger.Debug(fmt.Sprintf("[gRPC] Link storage doesn't contain hash \"%s\"", r.Hash))
		return nil, status.Error(codes.NotFound, "Link storage doesn't contain given hash")
	}

	if link.IsDeleted {
		application.logger.Debug(fmt.Sprintf("[gRPC] Client success get resource \"%s\", but it's was deleted", r.Hash))
		return nil, status.Error(codes.DataLoss, "Link was deleted")
	}

	application.logger.Debug(fmt.Sprintf("[gRPC] Client success get link \"%s\" by hash \"%s\"", link.Value, r.Hash))
	var response pb.GetLinkResponse
	response.Link = link.Value

	return &response, nil
}

// AddLinkBatch add batch of links by gRPC
func (application *Application) AddLinkBatch(ctx context.Context, r *pb.AddLinkBatchRequest) (*pb.AddLinkBatchResponse, error) {
	userUUIDCtxValue := ctx.Value(app.ContextUserUUIDKey)
	if userUUIDCtxValue == nil {
		application.logger.Error("[gRPC] No user UUID assigned")
		return nil, status.Error(codes.PermissionDenied, "No user UUID assigned")
	}

	userUUID := userUUIDCtxValue.(app.UserUUID)

	innerKWRows := make([]linkstorage.KeyValueRow, len(r.Links))
	expectedResult := make([]*pb.CorrelationHash, len(r.Links))
	for i, item := range r.Links {
		if !validateURL([]byte(item.Link)) {
			msg := fmt.Sprintf("[gRPC] Client sends invalid URL \"%s\" in batch item %s.", item.Link, item.Id)
			application.logger.Debug(msg)
			return nil, status.Errorf(codes.InvalidArgument, "Invalid link '%s' on ID %s", item.Link, item.Id)
		}

		hash, err := encoder.EncodeURL(item.Link)
		if err != nil {
			msg := fmt.Sprintf("[gRPC] Error while hashing URL \"%s\": %s.", item.Link, err.Error())
			application.logger.Debug(msg)
			return nil, status.Errorf(codes.Internal, "Error while hashing URL '%s' on ID %s", item.Link, item.Id)
		}

		innerKWRows[i] = linkstorage.KeyValueRow{
			Key:       hash,
			Value:     item.Link,
			OwnerUUID: string(userUUID),
		}

		expectedResult[i] = &pb.CorrelationHash{
			Id:   item.Id,
			Hash: application.buildTokenTail() + hash,
		}
	}

	application.loadLinkBatchBackground(innerKWRows)
	var response pb.AddLinkBatchResponse
	response.Hashes = expectedResult
	return &response, nil
}

// GetUserLinks get list of links owned by user by gRPC
func (application *Application) GetUserLinks(ctx context.Context, r *pb.GetUserLinksRequest) (*pb.GetUserLinksResponse, error) {
	userUUIDCtxValue := ctx.Value(app.ContextUserUUIDKey)
	if userUUIDCtxValue == nil {
		application.logger.Error("[gRPC] No user UUID assigned")
		return nil, status.Error(codes.PermissionDenied, "No user UUID assigned")
	}

	userUUID := userUUIDCtxValue.(app.UserUUID)

	result := make([]*pb.HashLink, 0)
	rows := application.storage.FindByUserUUID(ctx, string(userUUID))
	if len(rows) == 0 {
		return nil, status.Error(codes.NotFound, "No one link found for user")
	}

	for _, row := range rows {
		result = append(result, &pb.HashLink{
			OriginalUrl: row.Value,
			ShortUrl:    application.buildTokenTail() + row.Key,
		})
	}

	var response pb.GetUserLinksResponse
	response.Hl = result
	return &response, nil
}

// DeleteLink delete batch of user links by gRPC
func (application *Application) DeleteLink(ctx context.Context, r *pb.DeleteLinkRequest) (*pb.DeleteLinkResponse, error) {
	userUUIDCtxValue := ctx.Value(app.ContextUserUUIDKey)
	if userUUIDCtxValue == nil {
		application.logger.Error("[gRPC] No user UUID assigned")
		return nil, status.Error(codes.PermissionDenied, "No user UUID assigned")
	}

	userUUID := userUUIDCtxValue.(app.UserUUID)

	go func(userUUID app.UserUUID) {
		deleteBatch := make([]linkstorage.Hash, 0)
		for _, val := range r.Links {
			deleteBatch = append(deleteBatch, linkstorage.Hash{
				Value:     val,
				OwnerUUID: string(userUUID),
			})
		}

		err := application.storage.DeleteBatch(context.TODO(), deleteBatch)
		if err != nil {
			application.logger.Error("[gRPC] Error while delete user links", zap.String("User UUID", string(userUUID)))
		}
	}(userUUID)

	var response pb.DeleteLinkResponse
	return &response, nil
}

// GetServiceStats collect service stats by gRPC
func (application *Application) GetServiceStats(ctx context.Context, r *pb.GetServiceStatsRequest) (*pb.GetServiceStatsResponse, error) {
	urlCount, err := application.storage.GetUniqueURLsCount(ctx)
	if err != nil {
		application.logger.Error("[gRPC] cannot load unique urls for stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot load unique urls for stats")
	}

	userCount, err := application.storage.GetUniqueUsersCount(ctx)
	if err != nil {
		application.logger.Error("[gRPC] cannot load unique users for stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot load unique users for stats")
	}

	var response pb.GetServiceStatsResponse
	response.UsersCount = uint32(userCount)
	response.LinksCount = uint32(urlCount)
	return &response, nil
}

// Возвращает прослойку для gRPC, подмешивающая регистрацию пользователя (в случае его отсутсвия) в контекст для указанных методов
func getRegisterInterceptor(acceptedMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		methodName := getMethodName(info.FullMethod)
		for _, val := range acceptedMethods {
			if val == methodName {
				if ctx.Value(app.ContextUserUUIDKey) == nil {
					ctx = context.WithValue(ctx, app.ContextUserUUIDKey, app.GenerateUserUUID())
				}
			}
		}

		resp, err := handler(ctx, req)

		return resp, err
	}
}

// Возвращает прослойку для gRPC, проверяющая UserUUID в контексте запроса для указанных методов, в случае отсуствия - разворачивает запрос с ошибкой
func getLoginInterceptor(acceptedMethods []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		methodName := getMethodName(info.FullMethod)
		for _, val := range acceptedMethods {
			if val == methodName {
				resp, err := handler(ctx, req)
				return resp, err
			}
		}

		return nil, status.Error(codes.Unauthenticated, "Request required UserUUID in context")
	}
}

func getMethodName(fullMethod string) string {
	strs := strings.Split(fullMethod, "/")
	return strs[len(strs)-1]
}

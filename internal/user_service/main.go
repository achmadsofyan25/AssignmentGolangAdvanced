package main

import (
	"context"
	"log"
	"net"
	"wallet_gateway/internal/user_service/model"
	pb "wallet_gateway/internal/user_service/protos"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	db *gorm.DB
}

func (u *UserServer) GetUser(c context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var getUser model.User
	if err := u.db.Find(&getUser, req.Id).Error; err != nil {
		log.Println("cant get user")
		return nil, err
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:   int32(getUser.ID),
			Name: getUser.Name,
		},
	}, nil
}

func (u *UserServer) CreateUser(c context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	createdUser := model.User{
		Name: req.Name,
	}
	err := u.db.Create(&createdUser).Error
	if err != nil {
		log.Println("error create user")
		return nil, err
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:   int32(createdUser.ID),
			Name: createdUser.Name,
		},
	}, nil
}

func main() {
	dsn := "postgresql://postgres:pepega90@localhost:5432/db_user_grpc"
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatalf("cant connect to database: %v", err)
	}

	DB.AutoMigrate(&model.User{})

	userServer := grpc.NewServer()
	pb.RegisterUserServiceServer(userServer, &UserServer{db: DB})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("run grpc user 50051")
	if err := userServer.Serve(lis); err != nil {
		log.Fatalf("failed to run user grpc service: %v", err)
	}
}

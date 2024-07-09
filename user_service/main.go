package main

import (
	"context"
	"log"
	"net"
	"user_service/model"
	pb "user_service/protos"
	walletPb "wallet_service/protos"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	db           *gorm.DB
	walletClient walletPb.WalletServiceClient
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

	_, err = u.walletClient.CreateWallet(context.Background(), &walletPb.WalletRequest{
		UserId: int32(createdUser.ID),
	})
	if err != nil {
		log.Println("error creating wallet for user:", err)
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

	walletConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to wallet service: %v", err)
	}
	defer walletConn.Close()

	walletClient := walletPb.NewWalletServiceClient(walletConn)

	userServer := grpc.NewServer()
	pb.RegisterUserServiceServer(userServer, &UserServer{db: DB, walletClient: walletClient})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("run grpc user 50051")
	if err := userServer.Serve(lis); err != nil {
		log.Fatalf("failed to run user grpc service: %v", err)
	}
}

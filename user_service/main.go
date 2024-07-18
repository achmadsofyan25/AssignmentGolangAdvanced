package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"
	"user_service/model"
	"user_service/pkg"
	pb "user_service/protos"
	walletPb "wallet_service/protos"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	db           *gorm.DB
	walletClient walletPb.WalletServiceClient
	rdb          *redis.Client
}

func (u *UserServer) GetUser(c context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var getUser model.User

	val, err := u.rdb.Get(c, "get_user"+string(req.Id)).Result()
	if err == nil {
		err := json.Unmarshal([]byte(val), &getUser)
		if err == nil {
			user_wallet, err := u.walletClient.GetWallet(c, &walletPb.GetWalletRequest{UserId: int32(getUser.ID)})
			if err != nil {
				log.Println("cant get user wallet")
				return nil, err
			}

			list_transaction, err := u.walletClient.GetTransactions(c, &walletPb.GetTransactionsRequest{UserId: int32(getUser.ID)})
			if err != nil {
				log.Println("cant get user transactions")
				return nil, err
			}

			var list_trans []*pb.TransactionUser

			for _, v := range list_transaction.Transactions {
				list_trans = append(list_trans, &pb.TransactionUser{
					Type:   v.Type,
					Amount: v.Amount,
				})
			}

			return &pb.GetUserResponse{
				Id:           int32(getUser.ID),
				Name:         getUser.Name,
				Balance:      int32(user_wallet.Wallet.GetBalance()),
				Transactions: list_trans,
			}, nil
		}
	}

	if err := u.db.Find(&getUser, req.Id).Error; err != nil {
		log.Println("cant get user")
		return nil, err
	}

	user_wallet, err := u.walletClient.GetWallet(c, &walletPb.GetWalletRequest{UserId: int32(getUser.ID)})
	if err != nil {
		log.Println("cant get user wallet")
		return nil, err
	}

	list_transaction, err := u.walletClient.GetTransactions(c, &walletPb.GetTransactionsRequest{UserId: int32(getUser.ID)})
	if err != nil {
		log.Println("cant get user transactions")
		return nil, err
	}

	var list_trans []*pb.TransactionUser

	for _, v := range list_transaction.Transactions {
		list_trans = append(list_trans, &pb.TransactionUser{
			Type:   v.Type,
			Amount: v.Amount,
		})
	}

	byteData, _ := json.Marshal(getUser)
	u.rdb.SetEx(c, "get_user"+string(req.Id), string(byteData), 60*time.Second)

	return &pb.GetUserResponse{
		Id:           int32(getUser.ID),
		Name:         getUser.Name,
		Balance:      int32(user_wallet.Wallet.GetBalance()),
		Transactions: list_trans,
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

	// redis
	rdb := pkg.ConnectRedis()
	defer rdb.Close()

	walletClient := walletPb.NewWalletServiceClient(walletConn)
	userServer := grpc.NewServer()
	pb.RegisterUserServiceServer(userServer, &UserServer{db: DB, walletClient: walletClient, rdb: rdb})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("run grpc user 50051")
	if err := userServer.Serve(lis); err != nil {
		log.Fatalf("failed to run user grpc service: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"testing"
	"user_service/model"
	pb "user_service/protos"
	walletPb "wallet_service/protos"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockWalletClient struct {
	walletPb.WalletServiceClient
	ctrl *gomock.Controller
}

func (m *mockWalletClient) GetWallet(ctx context.Context, in *walletPb.GetWalletRequest, opts ...grpc.CallOption) (*walletPb.GetWalletResponse, error) {
	return &walletPb.GetWalletResponse{
		Wallet: &walletPb.Wallet{
			Id:      1,
			UserId:  in.UserId,
			Balance: 100,
		},
	}, nil
}

func (m *mockWalletClient) GetTransactions(ctx context.Context, in *walletPb.GetTransactionsRequest, opts ...grpc.CallOption) (*walletPb.GetTransactionsResponse, error) {
	return &walletPb.GetTransactionsResponse{
		Transactions: []*walletPb.Transaction{
			{Type: "Top Up", Amount: 50},
			{Type: "Transfer", Amount: 30},
		},
	}, nil
}

func (m *mockWalletClient) CreateWallet(ctx context.Context, in *walletPb.WalletRequest, opts ...grpc.CallOption) (*walletPb.WalletResponse, error) {
	return &walletPb.WalletResponse{
		Wallet: &walletPb.Wallet{
			Id:      1,
			UserId:  in.UserId,
			Balance: 0,
		},
	}, nil
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open test database: %v", err)
	}

	db.AutoMigrate(&model.User{})
	return db
}

func TestGetUser(t *testing.T) {
	db := setupTestDB()
	walletClient := &mockWalletClient{}
	service := &UserServer{db: db, walletClient: walletClient}

	// Setup a user
	user := model.User{ID: 1, Name: "John Doe"}
	db.Create(&user)

	req := &pb.GetUserRequest{Id: 1}
	resp, err := service.GetUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(user.ID), resp.Id)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, int32(100), resp.Balance)
	assert.Len(t, resp.Transactions, 2)

}

func TestCreateUser(t *testing.T) {
	db := setupTestDB()
	walletClient := &mockWalletClient{}
	service := &UserServer{db: db, walletClient: walletClient}

	req := &pb.CreateUserRequest{Name: "John Doe"}
	resp, err := service.CreateUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.Equal(t, int32(1), resp.User.Id)
}

package main

import (
	"context"
	"log"
	"testing"
	"wallet_service/model"
	pb "wallet_service/protos"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open test database: %v", err)
	}

	db.AutoMigrate(&model.Wallet{}, &model.Transaction{})
	return db
}

func TestCreateWallet(t *testing.T) {
	db := setupTestDB()
	service := &WalletService{db: db}

	req := &pb.WalletRequest{UserId: 1}
	resp, err := service.CreateWallet(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.UserId, resp.Wallet.UserId)
	assert.Equal(t, float32(0), resp.Wallet.Balance)
}

func TestGetWallet(t *testing.T) {
	db := setupTestDB()
	service := &WalletService{db: db}

	// Setup a wallet
	wallet := model.Wallet{UserID: 1, Balance: 100}
	db.Create(&wallet)

	req := &pb.GetWalletRequest{UserId: 1}
	resp, err := service.GetWallet(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.UserId, resp.Wallet.UserId)
	assert.Equal(t, float32(wallet.Balance), resp.Wallet.Balance)
}

func TestTopUp(t *testing.T) {
	db := setupTestDB()
	service := &WalletService{db: db}

	// Setup a wallet
	wallet := model.Wallet{UserID: 1, Balance: 100}
	db.Create(&wallet)

	req := &pb.TopUpRequest{UserId: 1, Amount: 50}
	resp, err := service.TopUp(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, float32(150), resp.Wallet.Balance)
}

func TestTransfer(t *testing.T) {
	db := setupTestDB()
	service := &WalletService{db: db}

	// Setup wallets
	wallet1 := model.Wallet{UserID: 1, Balance: 100}
	wallet2 := model.Wallet{UserID: 2, Balance: 50}
	db.Create(&wallet1)
	db.Create(&wallet2)

	req := &pb.TransferRequest{FromUserId: 1, ToUserId: 2, Amount: 50}
	resp, err := service.Transfer(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, float32(50), resp.Wallet.Balance)

	// Check balances
	var updatedWallet1, updatedWallet2 model.Wallet
	db.Where("user_id = ?", 1).First(&updatedWallet1)
	db.Where("user_id = ?", 2).First(&updatedWallet2)

	assert.Equal(t, float64(50), updatedWallet1.Balance)
	assert.Equal(t, float64(100), updatedWallet2.Balance)
}

func TestGetTransactions(t *testing.T) {
	db := setupTestDB()
	service := &WalletService{db: db}

	// Setup a wallet and transactions
	wallet := model.Wallet{UserID: 1, Balance: 100}
	db.Create(&wallet)

	transaction1 := model.Transaction{UserID: 1, Type: "Top Up", Amount: 50}
	transaction2 := model.Transaction{UserID: 1, Type: "Transfer", Amount: 20}
	db.Create(&transaction1)
	db.Create(&transaction2)

	req := &pb.GetTransactionsRequest{UserId: 1}
	resp, err := service.GetTransactions(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Transactions, 2)
}

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"wallet_service/model"
	pb "wallet_service/protos"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WalletService struct {
	pb.UnimplementedWalletServiceServer
	db *gorm.DB
}

func (w *WalletService) CreateWallet(c context.Context, req *pb.WalletRequest) (*pb.WalletResponse, error) {
	wallet := model.Wallet{
		UserID:  uint(req.UserId),
		Balance: 0,
	}
	err := w.db.Create(&wallet).Error
	if err != nil {
		log.Println("error create wallet")
		return nil, err
	}

	return &pb.WalletResponse{
		Wallet: &pb.Wallet{
			Id:      int32(wallet.ID),
			UserId:  int32(wallet.UserID),
			Balance: float32(wallet.Balance),
		},
	}, nil
}

func (w *WalletService) GetWallet(c context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	var wallet model.Wallet
	if err := w.db.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		log.Printf("error get wallet: %v", err)
		return nil, err
	}

	return &pb.GetWalletResponse{
		Wallet: &pb.Wallet{
			Id:      int32(wallet.ID),
			UserId:  int32(wallet.UserID),
			Balance: float32(wallet.Balance),
		},
	}, nil
}

func (w *WalletService) TopUp(c context.Context, req *pb.TopUpRequest) (*pb.TopUpResponse, error) {
	tx := w.db.Begin()
	var wallet model.Wallet
	if err := tx.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	wallet.Balance += float64(req.Amount)
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	transaction := model.Transaction{
		UserID: wallet.UserID,
		Type:   "Top Up",
		Amount: float64(req.Amount),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &pb.TopUpResponse{
		Wallet: &pb.Wallet{
			Id:      int32(wallet.ID),
			UserId:  int32(wallet.UserID),
			Balance: float32(wallet.Balance),
		},
	}, nil
}

func (w *WalletService) Transfer(c context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	tx := w.db.Begin()

	var fromUserWallet, toUserWallet model.Wallet
	if err := tx.Where("user_id = ?", req.FromUserId).First(&fromUserWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Where("user_id = ?", req.ToUserId).First(&toUserWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if fromUserWallet.Balance < float64(req.Amount) {
		tx.Rollback()
		return nil, fmt.Errorf("amount lebih besar dari balance!")
	}

	fromUserWallet.Balance -= float64(req.Amount)
	toUserWallet.Balance += float64(req.Amount)

	if err := tx.Save(&fromUserWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Save(&toUserWallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	transaction := model.Transaction{
		UserID: fromUserWallet.UserID,
		Type:   "Transfer",
		Amount: float64(req.Amount),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &pb.TransferResponse{
		Wallet: &pb.Wallet{
			UserId:  int32(fromUserWallet.UserID),
			Balance: float32(fromUserWallet.Balance),
		},
	}, nil
}

func (w *WalletService) GetTransactions(c context.Context, req *pb.GetTransactionsRequest) (*pb.GetTransactionsResponse, error) {
	var listTransaction []model.Transaction
	err := w.db.Where("user_id = ?", req.UserId).Find(&listTransaction).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}

	listTrans := []*pb.Transaction{}
	for _, v := range listTransaction {
		listTrans = append(listTrans, &pb.Transaction{
			Id:     uint32(v.ID),
			UserId: uint32(v.UserID),
			Type:   v.Type,
			Amount: float32(v.Amount),
		})
	}

	return &pb.GetTransactionsResponse{
		Transactions: listTrans,
	}, nil
}

func main() {
	dsn := "postgresql://postgres:pepega90@localhost:5432/db_wallet_grpc"
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatalf("cant connect to database: %v", err)
	}

	DB.AutoMigrate(&model.Wallet{}, &model.Transaction{})

	walletService := grpc.NewServer()
	pb.RegisterWalletServiceServer(walletService, &WalletService{db: DB})

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("run grpc wallet 50052")
	if err := walletService.Serve(lis); err != nil {
		log.Fatalf("failed to run user grpc service: %v", err)
	}
}

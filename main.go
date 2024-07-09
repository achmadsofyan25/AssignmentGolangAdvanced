package main

import (
	"log"
	"wallet_gateway/handler"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer userConn.Close()

	walletConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer walletConn.Close()

	r := gin.Default()
	h := handler.NewHandler(userConn, walletConn)

	r.POST("/user", h.CreateUser)
	r.POST("/wallet/topup", h.TopUp)
	r.POST("/wallet/transfer", h.Transfer)
	r.GET("/wallet/transactions/:id", h.GetUserTransactionList)

	log.Println("gateway run on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

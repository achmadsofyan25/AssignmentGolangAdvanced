package handler

import (
	"net/http"
	"strconv"
	userPb "wallet_gateway/internal/user_service/protos"
	walletPb "wallet_gateway/internal/wallet_service/protos"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Handler struct {
	userClient   userPb.UserServiceClient
	walletClient walletPb.WalletServiceClient
}

func NewHandler(userConn, walletConn *grpc.ClientConn) *Handler {
	return &Handler{
		userClient:   userPb.NewUserServiceClient(userConn),
		walletClient: walletPb.NewWalletServiceClient(walletConn),
	}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req userPb.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdUser, err := h.userClient.CreateUser(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	walletUser := walletPb.WalletRequest{
		UserId:  createdUser.GetUser().Id,
		Balance: 0,
	}
	createdWalletUser, _ := h.walletClient.CreateWallet(c, &walletUser)
	c.JSON(http.StatusOK, gin.H{"data": map[string]any{
		"user":   createdUser.User,
		"wallet": createdWalletUser.Wallet,
	}})
}

func (h *Handler) TopUp(c *gin.Context) {
	var req walletPb.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.walletClient.TopUp(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Transfer(c *gin.Context) {
	var req walletPb.TransferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err := h.walletClient.Transfer(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully transfer"})

}

func (h *Handler) GetUserTransactionList(c *gin.Context) {
	idUser, _ := strconv.Atoi(c.Param("id"))
	res, err := h.walletClient.GetTransactions(c, &walletPb.GetTransactionsRequest{UserId: int32(idUser)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

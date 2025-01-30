package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewTokenResponse struct {
	AccessToken string `json:"access_token"`
	AccessTokenExpires time.Time `json:"access_token_expires"`
}

func (server *Server) renewToken(ctx *gin.Context) {
	var req renewTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Verify the refresh token
	payload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Get session
	session, err := server.store.GetSession(ctx, payload.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Check if the session is blocked
	if session.IsBlocked {
		err := errors.New("Token is blocked")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Create new access token
	accessToken, payloadAccessToken, err := server.tokenMaker.CreateToken(payload.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return the access token
	response := renewTokenResponse{
		AccessToken: accessToken,
		AccessTokenExpires: payloadAccessToken.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, response)
}
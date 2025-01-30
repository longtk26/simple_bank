package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type User struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt time.Time `json:"created_at"`
}

func createUserResponse(user db.User) User {
	return User{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,	
		Password: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
				case "unique_violation", "foreign_key_violation":
					ctx.JSON(http.StatusForbidden, errorResponse(err))
					return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := createUserResponse(user)

	ctx.JSON(http.StatusOK, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type loginUserResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	AccessToken string `json:"access_token"`
	AccessTokenExpires time.Time `json:"access_token_expires"`
	RefreshToken string `json:"refresh_token"`
	RefreshTokenExpires time.Time `json:"refresh_token_expires"`
	User User `json:"user"`
}
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fmt.Println("loginUserRequest: ", req)
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, payloadAccessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, payloadRefreshToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(
		ctx,
		db.CreateSessionParams{
			ID: payloadRefreshToken.ID,
			Username: user.Username,
			RefreshToken: refreshToken,
			ExpiresAt: payloadRefreshToken.ExpiredAt,
			UserAgent: ctx.Request.UserAgent(),
			ClientIpAddress: ctx.ClientIP(),
			IsBlocked: false,
		},
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID: session.ID,
		AccessToken: accessToken,
		AccessTokenExpires: payloadAccessToken.ExpiredAt,
		RefreshToken: refreshToken,
		RefreshTokenExpires: payloadRefreshToken.ExpiredAt,
		User: createUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}
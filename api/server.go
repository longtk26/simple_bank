package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/longtk26/simple_bank/db/sqlc"
)

type Server struct {
	store db.Store
	router *gin.Engine
}

// NewServer create a new server instance and setup router
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes to the router
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
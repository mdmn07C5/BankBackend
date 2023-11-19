package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/mdmn07C5/bank/db/sqlc"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store  *db.Store   // allows interaction with db when processing API requests from clients
	router *gin.Engine // routes each API request to the correct handler for processing
}

// NewServer creates a new HTTP server instance and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// TODO: add routes to router

	server.router = router
	return server
}

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

// Start runs HTTP server on specified address and starts listening to API requests
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

// NewServer creates a new HTTP server instance and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// TODO: add routes to router
	router.POST("/accounts", server.createAccount) // if multiple handlers are passed, the last one is the real handler - the rest are middleware
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

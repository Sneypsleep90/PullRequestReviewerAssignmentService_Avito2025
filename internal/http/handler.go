package http

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	CreateUser(g *gin.Context)
	GetUserByID(g *gin.Context)
}

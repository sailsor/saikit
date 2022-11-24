package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 健康检查
type PingController struct {
}

func (pc *PingController) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"code":    "0000",
	})
}

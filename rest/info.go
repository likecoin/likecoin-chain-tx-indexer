package rest

import (
	_ "embed"

	"github.com/gin-gonic/gin"
)

//go:embed commit_hash.txt
var commitHash string

type InfoResponse struct {
	CommitHash string `json:"commit_hash"`
}

func handleInfo(c *gin.Context) {
	c.JSON(200, InfoResponse{CommitHash: commitHash})
}

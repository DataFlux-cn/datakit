package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type ping struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

func apiPing(c *gin.Context) {
	ErrOK.HttpBody(c, &ping{Version: datakit.Version, Uptime: fmt.Sprintf("%v", time.Since(uptime))})
}

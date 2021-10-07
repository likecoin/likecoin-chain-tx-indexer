package rest

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

func Run(pool *pgxpool.Pool, listenAddr string, lcdEndpoint string) {
	lcdURL, err := url.Parse(lcdEndpoint)
	if err != nil {
		logger.L.Panicw("Cannot parse lcd URL", "lcd_endpoint", lcdEndpoint, "error", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(lcdURL)
	proxyHandler := func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	router := gin.New()
	router.GET("/*endpoint", func(c *gin.Context) {
		endpoint, ok := c.Params.Get("endpoint")
		if !ok {
			logger.L.Errorw("Gin router cannot get endpoint")
			c.AbortWithStatus(500)
			return
		}
		switch endpoint {
		case "/txs":
			handleAminoTxsSearch(c, pool)
		case "/cosmos/tx/v1beta1/txs":
			handleStargateTxsSearch(c, pool)
		default:
			proxyHandler(c)
		}
	})
	router.POST("/*endpoint", proxyHandler)
	router.Run(listenAddr)
}

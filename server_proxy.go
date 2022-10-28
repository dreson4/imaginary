package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProxyServerOptions struct {
	ImaginaryPort int
	ProxyPort     int
	Endpoint      string
}

func ServerProxy(opts ProxyServerOptions) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Any("/ping", func(context *gin.Context) {
		context.String(http.StatusOK, "pong")
	})
	router.Any("/image/:item", func(ctx *gin.Context) {
		item := ctx.Param("item")
		action, ok := ctx.GetQuery("action")
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "action is missing"})
			return
		}

		link := fmt.Sprintf("http://localhost:%d/%s", opts.ImaginaryPort, action)
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		forwardQuery := ctx.Request.URL.Query()
		forwardQuery.Add("url", fmt.Sprintf("%s/%s", opts.Endpoint, item))
		forwardQuery.Del("action")
		req.URL.RawQuery = forwardQuery.Encode()

		client := http.DefaultClient
		res, err := client.Do(req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer res.Body.Close()
		io.Copy(ctx.Writer, res.Body)
	})

	log.Println("Proxy server listening on ", opts.ProxyPort)
	router.Run(fmt.Sprintf(":%d", opts.ProxyPort))
}

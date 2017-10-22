package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/wrfly/api-exporter/exporter"
	"gopkg.in/urfave/cli.v2"
)

func exporterMiddleware(c *gin.Context) {
	t := time.Now()
	c.Next()
	ip, method, path := c.ClientIP(), c.Request.Method, c.Request.URL.Path
	latency := time.Since(t)
	status := c.Writer.Status()
	auditStr := fmt.Sprintf("IP: [%s] - Method: [%s] Path: [%s] - Status: [%d] Latency: [%v]",
		ip, method, path, status, latency)
	logrus.Info(auditStr)
	exporter.Collect(ip, method, path, status, latency)
}
func run(c *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	logrus.SetLevel(logrus.DebugLevel)

	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(exporterMiddleware)

	r.GET("/200", func(c *gin.Context) {
		c.String(200, "200")
	})
	r.GET("/401", func(c *gin.Context) {
		c.String(401, "401")
	})
	r.GET("/500", func(c *gin.Context) {
		c.String(500, "500")
	})

	rs := []string{}
	for _, routes := range r.Routes() {
		rs = append(rs, routes.Path)
	}
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, rs)
	})

	r.GET("/metrics", func(c *gin.Context) {
		prometheus.Handler().ServeHTTP(c.Writer, c.Request)
	})

	p := fmt.Sprintf(":%s", c.String("port"))
	return r.Run(p)
}

func main() {
	logrus.Debug("Hello?")
	expoter := cli.App{
		Name:        "gin-exorter",
		Usage:       "a test web server useing gin-exporter lib",
		HideHelp:    true,
		HideVersion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
				Usage: "port to bind",
			},
		},
		Action: run,
	}
	expoter.Run(os.Args)
}

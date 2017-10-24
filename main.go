package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/wrfly/api-exporter/exporter"
	"gopkg.in/urfave/cli.v2"
)

func run(c *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)
	// gin.DefaultWriter = ioutil.Discard
	logrus.SetLevel(logrus.DebugLevel)

	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(exporter.GinMiddleware)

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

	p := fmt.Sprintf("localhost:%s", c.String("port"))
	go func() {
		r.Run(p)
	}()
	logrus.Infof("Server starts at %s", p)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	logrus.Infof("Server stopped due to [%s]", <-sigChan)

	return nil
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

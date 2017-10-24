package exporter

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var allGaugeVec = map[string]interface{}{
	"requestNum": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_num",
			Help: "Request numbers in default latency",
		},
		[]string{"path"},
	),
	"requestFromIPNum": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_from_num",
			Help: "Request from IP numbers in default latency",
		},
		[]string{"ip", "path", "code"},
	),
	"statusNum": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_status_num",
			Help: "Request status numbers in default latency",
		},
		[]string{"code", "path"},
	),
	"latencyTotal": prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_latency_total",
			Help: "Total request latency in 1m",
		},
		[]string{"path"},
	),
}

var uptime = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "uptime",
		Help: "Uptime in seconds",
	},
)

func init() {
	const timerDuration = 1 * time.Minute
	for _, collector := range allGaugeVec {
		prometheus.Register(collector.(prometheus.Collector))
	}

	prometheus.Register(uptime)

	go func() {
		for {
			time.Sleep(timerDuration)
			logrus.Debug("reset metrics")
			for _, collector := range allGaugeVec {
				collector.(*prometheus.GaugeVec).Reset()
			}
		}
	}()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			uptime.Inc()
		}
	}()
}

func collect(ip, UA, method, path string, status int, latency time.Duration) {

	if strings.Contains(UA, "Prometheus") {
		for _, collector := range allGaugeVec {
			c := collector.(*prometheus.GaugeVec)
			c.Reset()
		}
		return
	}

	for name, collector := range allGaugeVec {
		c := collector.(*prometheus.GaugeVec)
		switch name {
		case "requestNum":
			c.WithLabelValues(path).Inc()
		case "statusNum":
			c.WithLabelValues(strconv.Itoa(status), path).Inc()
		case "latencyTotal":
			c.WithLabelValues(path).Add(float64(latency))
		case "requestFromIPNum":
			c.WithLabelValues(ip, path, strconv.Itoa(status)).Inc()
		}
	}

	return
}

func GinMiddleware(c *gin.Context) {
	t := time.Now()
	c.Next()
	go func(c *gin.Context) {
		ip, UA, method, path := c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.URL.Path
		latency := time.Since(t)
		status := c.Writer.Status()
		collect(ip, UA, method, path, status, latency)
	}(c.Copy())
}

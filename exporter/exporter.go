package exporter

import (
	"strconv"
	"time"

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

type pathInfo struct {
	requestNum   int
	statusNum    map[string]int
	latencyTotal time.Time
}

type Exporter struct {
	paths    map[string]pathInfo
	ipNum    map[string]string
	startAt  time.Time
	duration time.Duration
}

func New(routes []string, duration time.Duration) {
	e := &Exporter{
		startAt:  time.Now(),
		duration: time.Second,
		paths:    make(map[string]pathInfo, 0),
		ipNum:    make(map[string]string, 0),
	}

	for _, route := range routes {
		logrus.Debugf("get route: %s", route)
		e.paths[route] = pathInfo{
			requestNum: 0,
			statusNum:  make(map[string]int, 0),
		}
	}

}

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

func Collect(ip, method, path string, status int, latency time.Duration) error {

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

	return nil
}

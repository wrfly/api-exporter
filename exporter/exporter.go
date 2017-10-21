package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func Test() string {
	t := prometheus.DefaultRegisterer
	logrus.Debug(t)
	return "test"
}

func RegisterRoutes(routes []string) error {

	return nil
}

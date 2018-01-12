package util

import (
	"time"
	"net"
	"github.com/golang/glog"
	"fmt"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
)

type ServerMetrics struct {
	handler http.Handler
	prediction_resp *prometheus.HistogramVec
	http_resp *prometheus.HistogramVec
}

func NewMetrics() *ServerMetrics {
	predict := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "predict_millseconds",
		Help: "Time taken to predict labels for image",
	}, []string{"code"})

	http := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "page_resp_millseconds",
		Help: "Overall time taken to predict the image and print the image",
	}, []string{"code"})

	prometheus.Register(predict)
	prometheus.Register(http)

	return &ServerMetrics{
		prediction_resp: predict,
		http_resp: http,
		handler: prometheus.Handler(),
	}
}

func (m *ServerMetrics) AddPrediction(code int, du time.Duration) {
	m.prediction_resp.WithLabelValues(fmt.Sprintf("%d", code)).Observe(du.Seconds()*1000.0)
}

func (m *ServerMetrics) AddHttp(code int, du time.Duration) {
	m.http_resp.WithLabelValues(fmt.Sprintf("%d", code)).Observe(du.Seconds()*1000.0)
}

func (m *ServerMetrics) Handle(w http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(w, r)
}

func TimeTrack(start time.Time, name string) time.Duration{
	elapsed := time.Since(start)
	glog.V(2).Infof("%s took %s", name, elapsed)
	return elapsed
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("are you connected to the network?")
}

func StringToInt(path string) int {
	if len(path) < 1 {
		return 0
	}

	bbuf := []byte(path)

	result := 0
	for i := 0; i < len(bbuf); i ++ {
		result += int(bbuf[i])
	}

	return result
}
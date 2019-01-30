package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"math"
)

var humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "humidity_percent",
}, []string{"mac"})

var temperatureCGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "temperature_c",
}, []string{"mac"})

var temperatureFGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "temperature_f",
}, []string{"mac"})

var batteryGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "battery_mv",
}, []string{"mac"})

var pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "pressure_kpa",
}, []string{"mac"})

var rssiGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "rssi",
}, []string{"mac"})

var accelerationXGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_x_g",
}, []string{"mac"})

var accelerationYGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_y_g",
}, []string{"mac"})

var accelerationZGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_z_g",
}, []string{"mac"})

func startMetrics() {
	go httpServ()
}

func httpServ() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func recordMetrics(mac string, rssi int16, data RuuviData) {
	tempC := float64(data.Temp) + (float64(data.TempDec) / 100)
	tempF := tempC * 9 / 5 + 32
	tempF = math.Round(tempF * 100) / 100
	humidityGauge.WithLabelValues(mac).Set(float64(data.Humidity))
	temperatureCGauge.WithLabelValues(mac).Set(tempC)
	temperatureFGauge.WithLabelValues(mac).Set(tempF)
	pressureGauge.WithLabelValues(mac).Set(float64(data.Pressure) / 1000)
	accelerationXGauge.WithLabelValues(mac).Set(float64(data.AccelerationX) / 1000)
	accelerationYGauge.WithLabelValues(mac).Set(float64(data.AccelerationY) / 1000)
	accelerationZGauge.WithLabelValues(mac).Set(float64(data.AccelerationZ) / 1000)
	batteryGauge.WithLabelValues(mac).Set(float64(data.Battery))
	rssiGauge.WithLabelValues(mac).Set(float64(rssi))
}

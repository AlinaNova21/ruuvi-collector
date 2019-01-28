package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	verbose = flag.Bool("v", false, "verbose")
)

type RuuviData struct {
	Id            int16
	Ver           int8
	Humidity      uint8
	Temp          int8
	TempDec       uint8
	Pressure      uint16
	AccelerationX int16
	AccelerationY int16
	AccelerationZ int16
	Battery       uint16
}

type RuuviNode struct {
	Mac  string
	Rssi int
	Data RuuviData
}

var nodes map[string]RuuviNode

var humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "humidity_percent",
}, []string{"mac"})
var temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "temperature_c",
}, []string{"mac"})
var batteryGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "battery_mv",
}, []string{"mac"})
var pressureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "pressure_pa",
}, []string{"mac"})
var rssiGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "rssi",
}, []string{"mac"})
var accelerationXGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_x",
}, []string{"mac"})
var accelerationYGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_y",
}, []string{"mac"})
var accelerationZGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ruuvi",
	Name:      "acceleration_z",
}, []string{"mac"})

func main() {
	flag.Parse()
	go httpServ()

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	err = ble.Scan(ctx, true, advHandler, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func recordMetrics(mac string, rssi int, data RuuviData) {
	humidityGauge.WithLabelValues(mac).Set(float64(data.Humidity))
	temperatureGauge.WithLabelValues(mac).Set(float64(data.Temp) + (float64(data.TempDec) / 100))
	pressureGauge.WithLabelValues(mac).Set(float64(data.Pressure))
	accelerationXGauge.WithLabelValues(mac).Set(float64(data.AccelerationX))
	accelerationYGauge.WithLabelValues(mac).Set(float64(data.AccelerationY))
	accelerationZGauge.WithLabelValues(mac).Set(float64(data.AccelerationZ))
	batteryGauge.WithLabelValues(mac).Set(float64(data.Battery))
	rssiGauge.WithLabelValues(mac).Set(float64(rssi))
}

func httpServ() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func advHandler(a ble.Advertisement) {
	md := a.ManufacturerData()
	if len(md) == 0 || md[0] != 0x99 && md[1] != 0x04 {
		return
	}
	data := RuuviData{}
	r := bytes.NewReader(md)
	_ = binary.Read(r, binary.BigEndian, &data)
	if *verbose {
		fmt.Printf("[%s] %3ddb\n", a.Addr(), a.RSSI())
		fmt.Printf("Version: %d\nHumidity: %d\nTemp: %d.%2dC\nPressure: %d Pa \nAcceleration: %d %d %d\nBattery: %dmv\n", data.Ver, data.Humidity, data.Temp, data.TempDec, data.Pressure+50000, data.AccelerationX, data.AccelerationY, data.AccelerationZ, data.Battery)
	}

	recordMetrics(fmt.Sprintf("%s", a.Addr()), a.RSSI(), data)
}

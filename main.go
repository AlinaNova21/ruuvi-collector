package main

import (
	"bytes"
	"encoding/binary"
	"flag"

	"os"

	"github.com/muka/go-bluetooth/api"
//	"github.com/muka/go-bluetooth/bluez/profile"
	"github.com/muka/go-bluetooth/emitter"
	"github.com/muka/go-bluetooth/linux"
	log "github.com/sirupsen/logrus"

)

var (
	verbose = flag.Bool("v", false, "verbose")
	logLevel = log.DebugLevel
	adapterID = flag.String("a", "hci0", "Adapter")
)

var nodes map[string]RuuviNode


func main() {
	flag.Parse()
	log.SetLevel(logLevel)

	startMetrics()

	//clean up connection on exit
	defer api.Exit()

	log.Debugf("Reset bluetooth device")
	linux.NewBtMgmt(*adapterID).Reset()
	var err error
	
	devices, err := api.GetDevices()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Infof("Cached devices:")
	for _, dev := range devices {
		handleDevice(&dev)
	}
	log.Infof("Discovered devices:")
	err = discoverDevices(*adapterID)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	select {}
}
func discoverDevices(adapterID string) error {

	err := api.StartDiscovery()
	if err != nil {
		return err
	}

	log.Debugf("Started discovery")
	err = api.On("discovery", emitter.NewCallback(func(ev emitter.Event) {
		discoveryEvent := ev.GetData().(api.DiscoveredDeviceEvent)
		dev := discoveryEvent.Device
		handleDevice(dev)
	}))
	return err
}

func handleDevice(dev *api.Device) {
	if dev == nil {
		return
	}
	props, err := dev.GetProperties()
  if err != nil {
		log.Errorf("%s: Failed to get properties: %s", dev.Path, err.Error())
		return
	}
	if _, ok := props.ManufacturerData[0x0499]; ok {
		parseAndRecord(dev)
		dev.On("changed", emitter.NewCallback(func(ev emitter.Event) {
			changedEvent := ev.GetData().(api.PropertyChangedEvent)
			parseAndRecord(changedEvent.Device)
		}))
	}
}

func parseAndRecord(dev *api.Device) error {
//func parseAndRecord(props *profile.Device1Properties) {
	props, err := dev.GetProperties()
	if err != nil {
		return err
	}
	rawData := props.ManufacturerData[0x0499].Value().([]byte)
	data := RuuviData{}
	r := bytes.NewReader(rawData)
	_ = binary.Read(r, binary.BigEndian, &data)
	if *verbose {
		log.Infof("[%s] %3ddb\n", props.Address, props.RSSI)
		log.Infof("md=%v", rawData)
		log.Infof("ver=%d humidity=%d tempc=%d.%2d pressure=%dPa acc=(%d %d %d) battery=%dmv", data.Ver, data.Humidity, data.Temp, data.TempDec, data.Pressure + 50000, data.AccelerationX, data.AccelerationY, data.AccelerationZ, data.Battery)
	}
	recordMetrics(props.Address, props.RSSI, data)
	return nil
}

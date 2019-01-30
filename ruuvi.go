package main

type RuuviData struct {
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

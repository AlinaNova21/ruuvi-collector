package main

import "encoding/binary"

type RuuviData struct {
	Ver           int8
	Humidity      uint8
	Temp          float64
	Pressure      uint16
	AccelerationX int16
	AccelerationY int16
	AccelerationZ int16
	Battery       uint16
}

func (r *RuuviData) FromBytes(raw []byte) {
  r.Ver = int8(raw[0])
	r.Humidity = uint8(raw[1])
	r.Temp = float64(raw[2] & 0x7F) + (float64(raw[3]) / 100)
	if raw[2] & 0x80 == 0x80 {
		r.Temp *= -1
	}
	r.Pressure = uint16(binary.BigEndian.Uint16(raw[4:6]))
	r.AccelerationX = int16(binary.BigEndian.Uint16(raw[6:8]))
	r.AccelerationY = int16(binary.BigEndian.Uint16(raw[8:10]))
	r.AccelerationZ = int16(binary.BigEndian.Uint16(raw[10:12]))
	r.Battery = binary.BigEndian.Uint16(raw[12:14])
}

type RuuviNode struct {
	Mac  string
	Rssi int
	Data RuuviData
}

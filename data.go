package main

import (
	"encoding/binary"
	"errors"
)

type Data struct {
	Address     string
	Temperature float32
	Humidity    uint8
	Battery     uint8
	// BatteryVoltage in millivolts
	BatteryVoltage uint16
}

func parseData(address string, b []byte) (*Data, error) {
	if len(b) < 12 {
		return nil, errors.New("malformed data bytes")
	}

	temp := int16(binary.BigEndian.Uint16(b[6:8]))
	hum := uint8(b[8])
	bat := uint8(b[9])
	vol := binary.BigEndian.Uint16(b[10:12])

	d := Data{
		Address:        address,
		Temperature:    float32(temp) / 10,
		Humidity:       hum,
		Battery:        bat,
		BatteryVoltage: vol,
	}

	return &d, nil
}

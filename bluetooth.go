package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slog"
	"tinygo.org/x/bluetooth"
)

const (
	//stadiaAddress = "A4:C1:38:CC:66:F8"
	//stadiaAddress = "A4:C1:38:3F:3E:EB"
	serviceUUID = "0000181A"
)

func FetchData(ctx context.Context, adapter *bluetooth.Adapter, devices []string, out chan *Data) error {
	if err := adapter.Enable(); err != nil {
		return fmt.Errorf("bluetooth: failed to enable adapter: %w", err)
	}

	errChan := make(chan error)
	scanResultChan := make(chan bluetooth.ScanResult)

	go func(errChan chan<- error) {
		err := adapter.Scan(scanCallback(scanResultChan, devices))
		if err != nil {
			errChan <- err
		}
	}(errChan)

	var err error

	counter := 0

	for {
		select {
		case <-ctx.Done():
			adapter.StopScan()
			return ctx.Err()
		case err = <-errChan:
			return err
		case scanRes := <-scanResultChan:
			bytes, err := parseServiceData(scanRes.ServiceData())
			if err != nil {
				return err
			}
			data, err := parseData(scanRes.Address.String(), bytes)
			if err != nil {
				return fmt.Errorf("cannot parse data: %w", err)
			}

			out <- data
			counter++
			if counter >= len(devices) {
				adapter.StopScan()
				close(out)
				return nil
			}
		}
	}
}

func scanCallback(resultChan chan<- bluetooth.ScanResult, devices []string) func(*bluetooth.Adapter, bluetooth.ScanResult) {
	return func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		found := make(map[string]bool)
		for _, d := range devices {
			found[d] = false
		}
		for _, d := range devices {
			slog.Debug("scanned device", "address", device.Address.String())
			if isTarget(&device, d) && !found[d] {
				found[d] = true
				slog.Info("target device found", "address", device.Address.String())
				resultChan <- device
			}
		}
	}
}

func isTarget(device *bluetooth.ScanResult, d string) bool {
	return device.Address.String() == d
}

func parseServiceData(serviceData []bluetooth.ServiceDataElement) ([]byte, error) {
	for _, sd := range serviceData {
		if matchUUID(sd.UUID, serviceUUID) {
			return sd.Data, nil

		}
	}
	return nil, errors.New("service data not found")
}

func matchUUID(charUUID bluetooth.UUID, targetUUID string) bool {
	return strings.HasPrefix(strings.ToUpper(charUUID.String()), targetUUID)
}

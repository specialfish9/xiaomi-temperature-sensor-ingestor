package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"tinygo.org/x/bluetooth"
)

func readData(dbConn driver.Conn, devices []string) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("Reading data from BLE devices", "device", devices)

	dataChan := make(chan *Data, len(devices))

	err := FetchData(bluetooth.DefaultAdapter, devices, dataChan)
	if err != nil {
		return fmt.Errorf("cannot fetch data: %w", err)
	}

	for d := range dataChan {
		slog.Info("Data fetched from BLE device", "device", d.Address, "temp", d.Temperature, "hum", d.Humidity, "bat", d.Battery, "volt", d.BatteryVoltage)

		if err := SaveData(context.Background(), dbConn, d); err != nil {
			err = errors.Join(err, fmt.Errorf("cannot save data to db: %w", err))
		}

		slog.Info("Data saved", "device", d.Address)
	}

	return err
}

func main() {
	fmt.Println("=========================")
	fmt.Println("XIAOMI BLE SENSOR READER")
	fmt.Println("=========================")

	cfg, err := NewConfig("config.conf")
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	slog.Debug("Config loaded")

	dbConn, err := Connect(cfg.DBAddress, cfg.DBName, cfg.DBUser, cfg.DBPassword, false)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	slog.Info("Connected to database")

	if err := Init(context.Background(), dbConn); err != nil {
		log.Fatalf("Cannot initialize database: %v", err)
	}

	slog.Info("Database initialized")

	devices := strings.Split(cfg.Devices, ",")

	ticker := time.NewTicker(time.Duration(cfg.Tick) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := readData(dbConn, devices); err != nil {
				slog.Error("Error reading data", "error", err)
			}
		}
	}

}

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect(address string, db string, username string, password string, debug bool) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: db,
			Username: username,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
		Debug:       debug, // prints queries for debugging
	})
	if err != nil {
		return nil, fmt.Errorf("db: cannot connect to db: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("db: cannot ping db: %w", err)
	}

	return conn, nil
}

func Init(ctx context.Context, conn driver.Conn) error {
	ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
		"max_execution_time": 60,
	}))

	// Create the queries table if it doesn't exist
	queries := []string{
		`CREATE DATABASE IF NOT EXISTS xiaomi;`,
		`
		CREATE TABLE IF NOT EXISTS data (
			address String,
		  temperature Float32,
			humidity UInt8,
			battery UInt8,
			battery_voltage UInt16,
			timestamp DateTime,
		) ENGINE = MergeTree() 
			ORDER BY (timestamp);
		`,
	}

	for i, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("db: cannot create initial table (%d): %w", i, err)
		}
	}

	return nil
}

func SaveData(ctx context.Context, conn driver.Conn, d *Data) error {
	err := conn.Exec(ctx, `
    INSERT INTO data (address, temperature, humidity, battery, battery_voltage, timestamp)
    VALUES (?, ?, ?, ?, ?, ?)
    `,
		d.Address,
		d.Temperature,
		d.Humidity,
		d.Battery,
		d.BatteryVoltage,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("db: cannot save data: %w", err)
	}

	return nil
}

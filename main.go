package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/joho/godotenv"
	"github.com/leogr/obdflux/pkg/obd"
	"github.com/rzetterberg/elmobd"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serialPath := flag.String(
		"serial",
		os.Getenv("SERIAL_DEVICE"),
		"Path to the serial device to use",
	)

	debug := flag.Bool(
		"debug",
		false,
		"Enable debug",
	)

	test := flag.Bool(
		"test",
		false,
		"Use test device",
	)

	flag.Parse()

	// Make OBD device
	dev, err := getDevice(*serialPath, *debug, *test)
	if err != nil {
		fmt.Println("Failed to create new device: ", err)
		return
	}

	// Get OBD device version
	version, err := dev.GetVersion()
	if err != nil {
		fmt.Println("Failed to get version: ", err)
		return
	}
	fmt.Println("Device version: ", version)

	// Get supported commands
	commands, err := getCommands(dev)
	if err != nil {
		fmt.Println("Failed to get supported commands: ", err)
		return
	}
	fmt.Printf("%d commands supported:\n", len(commands))
	for _, cmd := range commands {
		fmt.Println(" - ", cmd.Key())
	}

	// Make DB writer
	db, err := NewDBClient(client.HTTPConfig{
		Addr:     os.Getenv("DB_ADDR"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}, os.Getenv("DB_DATABASE"), os.Getenv("DB_MEASUREMENT"))
	if err != nil {
		fmt.Println("Failed to connect to InfluxDB: ", err)
		return
	}
	defer db.Close()

	// Make signals channels for graceful exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Starting...")

	// Make OBD scanner and start collecting data
	scanner := obd.NewScanner(dev, commands, time.Microsecond*100, 100)
	c := scanner.C()

	for {
		select {
		case ms := <-c:
			if *debug {
				fmt.Println(ms)
			}
			if err := db.Write(ms); err != nil {
				fmt.Println("Writing error: ", err)
			}
		case <-sigs:
			fmt.Println("Exiting...")
			scanner.Stop()
			return
		}
	}
}

func getDevice(devicePath string, debug bool, testDev bool) (*elmobd.Device, error) {
	if testDev {
		return elmobd.NewTestDevice(devicePath, debug)
	}
	return elmobd.NewDevice(devicePath, debug)
}

func getCommands(dev *elmobd.Device) ([]elmobd.OBDCommand, error) {
	supported, err := dev.CheckSupportedCommands()
	if err != nil {
		return nil, err
	}
	allCommands := elmobd.GetSensorCommands()
	commands := supported.FilterSupported(allCommands)
	return commands, nil
}

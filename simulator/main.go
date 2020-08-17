// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/tbrandon/mbserver"
)

const (
	defaultDevicePort      = 1502
	host                   = "0.0.0.0"
	startingPortEnvName    = "STARTING_PORT"
	simulatorNumberEnvName = "SIMULATOR_NUMBER"
	defaultStartingPort    = 2000
	defaultSimulatorNumber = 1000
)

var devices []*mbserver.Server
var mutex = &sync.Mutex{}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	_, scalabilityTestMode := os.LookupEnv(startingPortEnvName)

	if scalabilityTestMode {
		err := createScalabilityTestSimulators()
		if err != nil {
			log.Fatalf("Unalbe to create simulators for scalability test, %v. \n", err)
		}

		defer func() {
			for _, device := range devices {
				device.Close()
			}
			log.Printf("Close %d mock devices. \n", len(devices))
		}()

	} else {
		err := createMockDevice(defaultDevicePort)
		if err != nil {
			log.Fatalf("Fail to create simulator with port %d. \n", defaultDevicePort)
		}
	}

	<-c
	log.Println("Modbus simulator shutdown.")
}

func createMockDevice(port int) error {
	device := mbserver.NewServer()
	url := fmt.Sprintf("%s:%d", host, port)
	if err := device.ListenTCP(url); err != nil {
		log.Printf("Failed to start the Modbus TCP server as mock device, %v\n", err)
		return err
	}
	devices = append(devices, device)
	log.Printf("Start up a Modbus mock device with address %s \n", url)
	return nil
}

func scaleDevice(startingPort int, simulatorNumber int) (scaledDevicePorts []int, err error) {
	log.Printf("Create simulator, startingPort is %d, simulatorNumber is %d \n", startingPort, simulatorNumber)
	count := 0
	for count < simulatorNumber {
		err := createMockDevice(startingPort)
		if err != nil {
			return nil, err
		}
		scaledDevicePorts = append(scaledDevicePorts, startingPort)
		startingPort++
		count++
	}
	return scaledDevicePorts, nil
}

func createScalabilityTestSimulators() error {
	var err error
	startingPort := defaultStartingPort
	simulatorNumber := defaultSimulatorNumber
	startingPortEnv, ok := os.LookupEnv(startingPortEnvName)
	if ok {
		startingPort, err = strconv.Atoi(startingPortEnv)
		if err != nil {
			return fmt.Errorf("fail to parse STARTING_PORT %s. \n", startingPortEnv)
		}
	}
	simulatorNumberEnv, ok := os.LookupEnv(simulatorNumberEnvName)
	if ok {
		simulatorNumber, err = strconv.Atoi(simulatorNumberEnv)
		if err != nil {
			return fmt.Errorf("fail to parse SIMULATOR_NUMBER %s. \n", simulatorNumberEnv)
		}
	}

	_, err = scaleDevice(startingPort, simulatorNumber)
	if err != nil {
		return fmt.Errorf("fail to scale %d simulators. \n", simulatorNumber)
	}
	return nil
}

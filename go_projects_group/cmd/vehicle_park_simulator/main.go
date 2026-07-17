package main

import (
	"encoding/json"
	"fleet-app-gr/internal/payload"
	"fleet-app-gr/internal/vechpath"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type config struct {
	url      string
	user     string
	password string
}

func main() {
	_ = godotenv.Load() // для локального запуска

	mqttConfigs := config{
		url:      os.Getenv("MQTT_BROKER_URL"),
		user:     os.Getenv("MQTT_USER"),
		password: os.Getenv("MQTT_PASS"),
	}

	fmt.Printf("MQTT конфиг %+v", mqttConfigs)

	// Базовые координаты центрального гаража
	baseLat := 55.7485
	baseLon := 37.6085
	vehicleCount := 45

	var wg sync.WaitGroup
	log.Printf("Старт симуляции %d количества ТС", vehicleCount)

	engineOffIndices := map[int]bool{
		8: true, 9: true, 11: true, 12: true,
		22: true, 23: true, 33: true, 34: true,
	}

	for i := 1; i <= vehicleCount; i++ {
		wg.Add(1)

		vehicleType, fuelType := getVehicleSpecs(i)
		if vehicleType == "unknown" || fuelType == "unknown" {
			continue
		}

		engineStatus := "on"
		if engineOffIndices[i] {
			engineStatus = "off"
		}

		vehicleInfo := payload.VehicleInfo{
			ID:           i,
			StartLat:     baseLat + (float64(i)-1)*0.003,
			StartLon:     baseLon + (float64(i)-1)*0.003,
			VehicleType:  vehicleType,
			FuelType:     fuelType,
			EngineStatus: engineStatus}

		go func(vehicleInfo payload.VehicleInfo, mqttConfigs config) {
			defer wg.Done()

			runVehicle(vehicleInfo, mqttConfigs)
		}(vehicleInfo, mqttConfigs)
	}

	wg.Wait()
}

/*
1..8   tractor
9..20  forklift
21..29 robot
30..45 cart
*/
func getVehicleSpecs(i int) (vehicleType string, fuelType string) {
	switch {
	case i >= 1 && i <= 8:
		vehicleType = "tractor"
		fuelType = "diesel"
	case i >= 9 && i <= 20:
		vehicleType = "forklift"
		fuelType = "electric"
	case i >= 21 && i <= 29:
		vehicleType = "robot"
		fuelType = "electric"
	case i >= 30 && i <= 45:
		vehicleType = "cart"
		fuelType = "electric"
	default:
		vehicleType = "unknown"
		fuelType = "unknown"
	}
	return
}

func runVehicle(vehicleInfo payload.VehicleInfo, mqttConfigs config) {
	// индивидуальный сдвиг старта для каждой машины,
	// чтобы они не отправляли данные одновременно и алерты шли плавной волной.
	startupDelay := time.Duration(vehicleInfo.ID) * 2 * time.Second
	time.Sleep(startupDelay)

	options := mqtt.NewClientOptions().AddBroker(mqttConfigs.url)
	options.SetClientID(fmt.Sprintf("aynur_telemetry_gen_%d", vehicleInfo.ID))
	options.SetKeepAlive(2 * time.Second)
	options.SetPingTimeout(2 * time.Second)
	options.SetPassword(mqttConfigs.password)
	options.SetUsername(mqttConfigs.user)

	client := mqtt.NewClient(options)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Ошибка соединения к MQTT брокеру: %v", token.Error())
	}
	defer client.Disconnect(500)

	log.Printf("%d - %s - %s открыл соеднинение к MQTT брокеруr", vehicleInfo.ID, vehicleInfo.VehicleType, vehicleInfo.FuelType)

	path := vechpath.GenerateCircularPath(&vehicleInfo.StartLat, &vehicleInfo.StartLon)
	pointIndex := 0

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentCoord := path[pointIndex]
		pointIndex = (pointIndex + 1) % len(path)

		payload := payload.CreatePayload(vehicleInfo, currentCoord)
		jsonBytes, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			log.Fatalf("Ошибка сериализации JSON: %v", err)
			continue
		}
		topic := fmt.Sprintf("%s/%s/%d/telemetry", vehicleInfo.VehicleType, vehicleInfo.FuelType, vehicleInfo.ID)

		token := client.Publish(topic, 0, false, jsonBytes)
		token.Wait()
		if token.Error() != nil {
			log.Fatalf("Ошибка соединения к MQTT брокеру: %v", token.Error())
		} else {
			log.Printf("[ОПУБЛИКОВАНО] Топик: %s \n тело: %+v", topic, payload)
		}
	}
}

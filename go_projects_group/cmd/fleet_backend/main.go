package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	httpRequestsTotal         *prometheus.CounterVec
	httpDuration              *prometheus.HistogramVec
	telemetryPacketsProcessed *prometheus.CounterVec
	telemetryPacketsDropped   prometheus.Counter
}
type basePayload struct {
	VehicleType string `json:"vehicle_type"`
}

func registerMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		// 1. RED-метод: Счётчик HTTP-запросов с разделением по методу и статус-коду
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fms_http_requests_total",
				Help: "Общее количество обработанных HTTP запросов к FMS API.",
			},
			[]string{"method", "status", "path"},
		),

		// 2. RED-метод: Гистограмма длительности обработки запросов
		httpDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fms_http_request_duration_seconds",
				Help:    "Время обработки HTTP запросов бэкендом FMS.",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5}, // Границы бакетов в секундах
			},
			[]string{"path"},
		),

		// 3. Бизнес-логика: Количество обработанных пакетов телеметрии транспорта
		telemetryPacketsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fms_telemetry_packets_processed_total",
				Help: "Количество успешно обработанных пакетов телеметрии из MQTT.",
			},
			[]string{"vehicle_type"},
		),

		// 4. Бизнес-логика: Количество отброшенных поврежденных пакетов
		telemetryPacketsDropped: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "fms_telemetry_packets_dropped_total",
				Help: "Количество отброшенных пакетов телеметрии из-за ошибок парсинга.",
			},
		),
	}
	reg.MustRegister(m.httpRequestsTotal)
	reg.MustRegister(m.httpDuration)
	reg.MustRegister(m.telemetryPacketsProcessed)
	reg.MustRegister(m.telemetryPacketsDropped)

	// 2. ВАЖНО: Возвращаем стандартные метрики процесса и Go в новый реестр
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	return m
}

func main() {
	_ = godotenv.Load() // for locl dev purpose

	reg := prometheus.NewRegistry()
	m := registerMetrics(reg)
	go initMqttClient(m)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	// Пример роута API для диспетчеров FMS
	mux.HandleFunc("/api/v1/fleet/status", fleetStatusHandler(m))

	url := os.Getenv("FLEET_BACKEND_URL")
	if url == "" {
		url = ":8080" // Фоллбэк на случай, если в .env забыли указать порт
	}

	log.Printf("Бэкенд запущен на %s\n", url)
	if err := http.ListenAndServe(url, mux); err != nil {
		log.Fatal(err)
	}

}

func initMqttClient(m *metrics) {
	opts := mqtt.NewClientOptions().AddBroker(os.Getenv("MQTT_BROKER_URL"))
	opts.SetClientID("fms_backend_client")
	opts.SetUsername(os.Getenv("MQTT_USER"))
	opts.SetPassword(os.Getenv("MQTT_PASS"))
	opts.SetCleanSession(true)

	var msghandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		var p basePayload

		if err := json.Unmarshal(msg.Payload(), &p); err != nil {
			log.Printf("[MQTT Error]: Ошибка сериализации сообщения - %s", err)
			m.telemetryPacketsDropped.Inc()
			return
		}

		m.telemetryPacketsProcessed.WithLabelValues(p.VehicleType).Inc()
		log.Printf("[API MQTT] Пакет успешно учтен для ТС типа: %s", p.VehicleType)
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("[MQTT FATAL] Не удалось подключиться к MQTT: %v. Повтор через 2с...", token.Error())
		time.Sleep(2 * time.Second)
		go initMqttClient(m)
		return
	}

	topicPattern := "+/+/+/telemetry"
	if token := client.Subscribe(topicPattern, 0, msghandler); token.Wait() && token.Error() != nil {
		log.Fatalf("[MQTT FATAL] Ошибка подписки на топики: %v", token.Error())
	}
	log.Printf("[MQTT SUCCESS] Бэкенд успешно подписался на топик %s", topicPattern)
}

func fleetStatusHandler(m *metrics) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		// Симуляция случайной задержки обработки запроса диспетчера
		duration := float64(rand.Intn(200)) / 1000.0
		time.Sleep(time.Duration(duration * float64(time.Second)))

		// Симулируем случайные ошибки 500 (10% вероятность)
		status := "200"
		if rand.Intn(10) == 0 {
			status = "500"
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Internal Server Error"}`))
		} else {
			w.Write([]byte(`{"status": "all systems nominal"}`))
		}

		// Запись RED-метрик
		m.httpRequestsTotal.WithLabelValues(r.Method, status, path).Inc()
		m.httpDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())
	}
}

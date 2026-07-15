package main

type TelemetryPayload struct {
	SchemaVersion int    `json:"schema_version"`
	VehicleID     int    `json:"vehicle_id"`
	VehicleType   string `json:"vehicle_type"`
	FuelType      string `json:"fuel_type"`
	Timestamp     int64  `json:"timestamp"`

	Events []Event `json:"events"`
}

type Event struct {
	EventType   string `json:"event_type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Timestamp   int64  `json:"timestamp"`
}

type MetricCommon struct {
	GpsLat       float64 `json:"gps_lat"`
	GpsLon       float64 `json:"gps_lon"`
	GpsAlt       float64 `json:"gps_alt"`
	SpeedKmh     float32 `json:"speed_kmh"`
	EngineStatus string  `json:"engine_status"`
}

type DieselTelemetryPayload struct {
	TelemetryPayload
	Metrics DieselMetrics `json:"metrics"`
}

type DieselMetrics struct {
	MetricCommon
	EngineRPM      int     `json:"engine_rpm"`
	FuelLevelPct   int     `json:"fuel_level_pct"`
	TempC          float32 `json:"temp_c"`
	OilPressureBar float32 `json:"oil_pressure_bar"`
	EngineHours    float32 `json:"engine_hours"`
}

type ElectricTelemetryPayload struct {
	TelemetryPayload
	Metrics ElectricMetrics `json:"metrics"`
}

type ElectricMetrics struct {
	MetricCommon
	BatterySocPct int     `json:"battery_soc_pct"`
	BatteryTempC  float32 `json:"battery_temp_c"`
	CurrentA      float32 `json:"current_a"`
	VoltageV      float32 `json:"voltage_v"`
}

type RobotTelemetryPayload struct {
	TelemetryPayload
	Metrics RobotMetrics `json:"metrics"`
}

type RobotMetrics struct {
	ElectricMetrics
	Mode             string  `json:"mode"`           // idle, human, teleop, supervis, autonom
	MissionStatus    string  `json:"mission_status"` // none, pause, run, complete, autonom
	MissionID        string  `json:"mission_id"`
	EstopStatus      string  `json:"estop_status"` // on/off
	RTKStatus        string  `json:"rtl_status"`   // fix, float, none
	SteeringAngleDeg float64 `json:"steering_angle_deg"`
	TempCpuC         float32 `json:"temp_cpu_c"`
	LteRssi          float32 `json:"lte_rssi"`
}

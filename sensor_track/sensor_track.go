package sensor_track

import (
	"encoding/json"
	"github.com/DrWrong/monica/config"
	"github.com/DrWrong/monica/log"
	"time"
)

var sensorLogger *log.MonicaLogger

type EventTracer struct {
	DistinctId string                 `json:"distinct_id"`
	OriginId   string                 `json:"original_id,omitempty"`
	Time       int64                  `json:"time"`
	Type       string                 `json:"type"`
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
}

func (t *EventTracer) Tracer() {
	if t.Time == 0 {
		t.Time = time.Now().UnixNano() / 1e6
	}
	if t.Type == "" {
		t.Type = "track"
	}
	record, _ := json.Marshal(t)
	sensorLogger.Infof("%s", record)
}

func InitSensorLogger() {
	logName := config.GlobalConfiger.String("sensor_logger")
	sensorLogger = log.GetLogger(logName)
}

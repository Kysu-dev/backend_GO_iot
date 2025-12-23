package mqtt

import (
	"sync"
	"time"
)

const defaultSensorBatchInterval = 5 * time.Second

type sensorCache struct {
	mu sync.Mutex

	temperature        float64
	temperaturePending bool
	temperatureKnown   bool

	humidity        float64
	humidityPending bool
	humidityKnown   bool

	light        int
	lightPending bool
	lightKnown   bool

	gas        int
	gasPending bool
	gasKnown   bool
}

func (h *MQTTHandler) setLatestTemperature(value float64) {
	h.sensorCache.mu.Lock()
	h.sensorCache.temperature = value
	h.sensorCache.temperaturePending = true
	h.sensorCache.temperatureKnown = true
	h.sensorCache.mu.Unlock()
}

func (h *MQTTHandler) setLatestHumidity(value float64) {
	h.sensorCache.mu.Lock()
	h.sensorCache.humidity = value
	h.sensorCache.humidityPending = true
	h.sensorCache.humidityKnown = true
	h.sensorCache.mu.Unlock()
}

func (h *MQTTHandler) setLatestLight(value int) {
	h.sensorCache.mu.Lock()
	h.sensorCache.light = value
	h.sensorCache.lightPending = true
	h.sensorCache.lightKnown = true
	h.sensorCache.mu.Unlock()
}

func (h *MQTTHandler) setLatestGas(value int) {
	h.sensorCache.mu.Lock()
	h.sensorCache.gas = value
	h.sensorCache.gasPending = true
	h.sensorCache.gasKnown = true
	h.sensorCache.mu.Unlock()
}

package queue

import (
	"sync"

	"github.com/enorith/framework/queue/contracts"
)

type WorkerConfig struct {
	Connection  string `yaml:"connection"`
	Concurrency int    `yaml:"concurrency"`
}

type Config struct {
	Connection     string                            `yaml:"connection" env:"QUEUE_CONNECTION" default:"mem"`
	RunningWorkers []string                          `yaml:"running_workers"`
	Workers        map[string]WorkerConfig           `yaml:"workers"`
	Connections    map[string]map[string]interface{} `yaml:"connections"`
}

type Manager struct {
	connections map[string]contracts.Connection
	workers     map[string]contracts.Worker
	m           sync.RWMutex
}

func (m *Manager) Run(workers ...string) {

}

func (m *Manager) GetConnection(connection string) (contracts.Connection, bool) {
	m.m.RLock()
	defer m.m.RUnlock()

	c, ok := m.connections[connection]

	return c, ok
}

func (m *Manager) GetWorker(worker string) (contracts.Worker, bool) {
	m.m.RLock()
	defer m.m.RUnlock()

	w, ok := m.workers[worker]

	return w, ok
}

package queue

import "sync"

var counterMutex sync.Mutex

func initMutex() {
	counterMutex = sync.Mutex{}
}

var CounterRedisRDB int

func updateRedisRDBCounter() {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	CounterRedisRDB++
}

var CounterRedisAOF int

func updateRedisAOFCounter() {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	CounterRedisAOF++
}

var CounterRabbitMQ int

func updateRabbitMQCounter() {
	counterMutex.Lock()
	CounterRabbitMQ++
	counterMutex.Unlock()
}

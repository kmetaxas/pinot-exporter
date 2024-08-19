package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

/*
A WorkerPool is Per-Pinot cluster
Its job is to collect metrics from all tables in the cluster.
Since there could be a lot of tables, we want to parallelize this task
*/
type CollectorWorkerPool struct {
	wg                 sync.WaitGroup
	controller         PinotControllerInterface
	incomingTablesChan <-chan []string
	tables             chan string
	semaphore          chan struct{}
	numWorkers         int
}

func NewCollectorWorkerPool(numWorkers int, controller PinotControllerInterface, incomingTablesChan <-chan []string) *CollectorWorkerPool {
	pool := CollectorWorkerPool{
		controller:         controller,
		incomingTablesChan: incomingTablesChan,
		numWorkers:         numWorkers,
		semaphore:          make(chan struct{}, numWorkers),
		tables:             make(chan string),
	}
	// Start workers
	for i := 1; i <= numWorkers; i++ {
		ctx := context.Background()
		pool.wg.Add(1)
		go worker(i, ctx, pool.tables, pool.controller, pool.semaphore, &pool.wg)
	}

	return &pool
}

func (c *CollectorWorkerPool) Close() {
	c.wg.Wait()
	close(c.tables)
}

// Receive table array updates
func (c *CollectorWorkerPool) SubscribeToTableUpdates(tables <-chan []string) {
	for newTables := range tables {
		logger.Debugf("Pool received []table update: %+v\n", newTables)
		for _, table := range newTables {
			c.tables <- table
		}

	}
}

// Worker function that fetches the metric from the REST API
func worker(id int, ctx context.Context, tables <-chan string, controller PinotControllerInterface, semaphore chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Infof("Started collector worker with id %d for pinot %s", id, controller)
	for table := range tables {
		logger.Debugf("worker %d consumed table update '%+v' from channel.", id, table)
		// Acquire semaphore
		semaphore <- struct{}{}

		go func(table string) {
			defer func() { <-semaphore }() // Release semaphore
			// Introduce random jitter (0 to 500 ms)
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			time.Sleep(jitter)
			logger.Debugf("Worker %d of (%s) collecting size for table %s", id, controller, table)
			size, err := controller.GetSizeForTable(ctx, table)
			if err != nil {
				logger.Errorf("Failed to get size for table %s with error %s\n", table, err)
				return
			}
			TableSizeBytes.WithLabelValues(table).Set(float64(size))
		}(table)
	}
	logger.Infof("Worker with id %d, that was monitoring %s is returning", id, controller)
}

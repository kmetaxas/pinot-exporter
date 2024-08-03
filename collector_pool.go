package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type CollectorWorkerPool struct {
	wg                 sync.WaitGroup
	controller         *PinotController
	incomingTablesChan <-chan []string
	tables             chan string
	semaphore          chan struct{}
	numWorkers         int
}

func NewCollectorWorkerPool(numWorkers int, controller *PinotController, incomingTablesChan <-chan []string) *CollectorWorkerPool {
	pool := CollectorWorkerPool{
		controller:         controller,
		incomingTablesChan: incomingTablesChan,
		numWorkers:         numWorkers,
		semaphore:          make(chan struct{}, numWorkers),
		tables:             make(chan string, 1),
	}
	// Semaphore to limit concurrent workers
	// Number of workers

	// Start workers
	for i := 1; i <= numWorkers; i++ {
		ctx := context.Background()
		pool.wg.Add(1)
		go worker(ctx, pool.tables, pool.controller, pool.semaphore, &pool.wg)
	}

	return &pool
}

func (c *CollectorWorkerPool) Close() {
	c.wg.Wait()
	close(c.tables)
}

// Receive table array updates
func (c *CollectorWorkerPool) SubscribeToTableUpdates(tables <-chan []string) {
	for {
		select {
		case newTables := <-tables:
			{
				fmt.Printf("Pool received []table update: %+v\n", newTables)
				for _, table := range newTables {
					c.tables <- table
				}
			}
		default:
		}
	}

}

// Worker function that fetches the metric from the REST API
func worker(ctx context.Context, tables <-chan string, controller *PinotController, semaphore chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for table := range tables {
		// Acquire semaphore
		semaphore <- struct{}{}

		go func(table string) {
			defer func() { <-semaphore }() // Release semaphore
			// Introduce random jitter (0 to 500 ms)
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			time.Sleep(jitter)
			size, err := controller.GetSizeForTable(ctx, table)
			if err != nil {
				fmt.Printf("Failed to get size for table %s with error %s\n", size, err)
				return
			}
			TableSizeBytes.WithLabelValues(table).Set(float64(size))
		}(table)
	}
}
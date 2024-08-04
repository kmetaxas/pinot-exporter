package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Our metrics
var (
	TableSizeBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pinotexporter_table_size_bytes",
		Help: "Table size in bytes",
	},
		[]string{"table"},
	)
)

/*
FanOut consumer receives table updates from the tables channel and
distributes messages to all interested recepients
*/

func tableFanOutConsumer(tables <-chan []string, tableCache TableCache, workerPool *CollectorWorkerPool) {
	// First setup the refresh listener using another channel that this goroutine will copy into
	tablesCopyForCache := make(chan []string, 1)
	tablesCopyForPool := make(chan []string, 1)
	go tableCache.TableRefreshChanListener(tablesCopyForCache)
	go workerPool.SubscribeToTableUpdates(tablesCopyForPool)

	for {
		select {
		case newTables := <-tables:
			{
				// push record into the copy for TableRefreshChanListener
				tablesCopyForCache <- newTables
				// Send each record into the
				tablesCopyForPool <- newTables
			}
		default:
		}
	}

}
func main() {

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	tables := make(chan []string, 1)
	var tableCache TableCache

	conf, err := NewConfigFromFile("testconfig.yaml")
	workerPool := NewCollectorWorkerPool(conf.MaxParallelCollectors, conf.PinotController, tables)
	defer workerPool.Close()

	ctx := context.Background()

	go refreshTableCache(ctx, conf.PinotController, conf.PollFrequencySeconds, tables)
	go tableFanOutConsumer(tables, tableCache, workerPool)

	if err != nil {
		panic(err)
	}
	fmt.Printf("conf = %+v\n", conf)

	for _, table := range tableCache.GetTables() {
		ctx = context.Background()
		size, err := conf.PinotController.GetSizeForTable(ctx, table)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Size for %s= %d", table, size)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", conf.ListenPort), nil)

}

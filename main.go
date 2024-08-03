package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Our metrics
var (
	TableSizeBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinotexporter_table_size_bytes",
		Help: "Table size in bytes",
	})
)

func main() {

	tables := make(chan []string, 1)

	var tableCache TableCache

	conf, err := NewConfigFromFile("testconfig.yaml")

	ctx := context.Background()
	go refreshTableCache(ctx, conf.PinotController, conf.PollFrequencySeconds, tables)
	go tableCache.TableRefreshChanListener(tables)

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
	http.ListenAndServe(":2112", nil)

}

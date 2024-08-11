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
	"go.uber.org/zap"
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

// yeah, yeah , this is a bad practice and we should pass logger explicitly everywhere..
var logger *zap.SugaredLogger

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
				// Inform the pool about table updates
				tablesCopyForPool <- newTables
			}
		default:
		}
	}

}
func main() {

	//setup logging
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger = zapLogger.Sugar()
	logger.Infof("Started logging")

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	conf, err := NewConfigFromFile("testconfig.yaml")
	logger.Debugf("conf = %+v\n", conf)
	err = conf.IsValid()
	if err != nil {
		panic(err)
	}
	// IF Direct mode
	if conf.Mode == "direct" {
		logger.Info("Starting on Direct mode")
		var tableCache TableCache
		tables := make(chan []string, 1)
		workerPool := NewCollectorWorkerPool(conf.MaxParallelCollectors, conf.PinotController, tables)
		defer workerPool.Close()

		ctx := context.Background() // TODO set a timeout

		go refreshTableCache(ctx, conf.PinotController, conf.PollFrequencySeconds, tables)
		go tableFanOutConsumer(tables, tableCache, workerPool)

		if err != nil {
			panic(err)
		}
	}

	// IF Kubernetes MODE
	if conf.Mode == "kubernetes" {
		/*
		   We need track know Pinot clusters and on each update:
		   -
		*/
		logger.Info("Starting on Kubernetes mode")
		kubeClient := NewKubePinotControllerCache(conf.ServiceDiscovery)
		pinotManager, err := NewPinotManager(conf.MaxParallelCollectors, conf.PollFrequencySeconds, kubeClient)
		if err != nil {
			logger.Errorf("Can't create new PinotManager because: %s", err)
			panic(err)
		}
		go pinotManager.refreshPinotsForever()

	}

	// Start serving metrics
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", conf.ListenPort), nil)

}

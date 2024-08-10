package main

import (
	"fmt"
	"time"
)

/*
Manages a series of Pinot clusters
handles management of discovery, mebmership , metrics collection
*/

type PinotManager struct {
	// known pinots.
	knownPinots map[string]PinotController
	// table caches per known endpoint
	tableCaches map[string]*TableCache
	workerPools map[string]*CollectorWorkerPool
	// channels to get Table updates from, for each pinot service endpoint (key)
	tableChannels map[string](chan []string)
	// Channels to get send table updates to Collector Pool for each pinot endpoint (key)
	poolTableNotifyChannels map[string](chan []string)
	numConnectorWorkers     int
	// Seconds
	refreshInteval int
	// kuberneted controller cache
	kubeCache *KubePinotControllerCache
}

func NewPinotManager(numWorkers int, refreshInteval int, kubeCache *KubePinotControllerCache) (*PinotManager, error) {
	// setup with defaults
	mgr := &PinotManager{
		knownPinots:             make(map[string]PinotController),
		tableCaches:             make(map[string]*TableCache),
		workerPools:             make(map[string]*CollectorWorkerPool),
		tableChannels:           make(map[string](chan []string)),
		poolTableNotifyChannels: make(map[string](chan []string)),
		kubeCache:               kubeCache,
		numConnectorWorkers:     numWorkers,
		refreshInteval:          refreshInteval,
	}
	// TODO some validation and sanity checks
	return mgr, nil
}

func (m *PinotManager) refreshPinotsForever() {
	// Refresh
	err := m.kubeCache.Connect()
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		// This call refreshes the internal cache of m.kubeCache but also returns the results to us
		endpoints := m.kubeCache.refreshPinotClustersList()
		// This is very naive. We should *add* and *remove* entries gracefully as each entry has associated workers and goroutines
		m.updateKnownPinotsCache(endpoints)
	}
}

/*
Updates known pinots.
Checks if a pinot in the arguments is:
- existing: Does nothing
- New: Adds a new TableCache and CollectorPool
- Deleted: Removes an existing TableCache and CollectorPool
*/
func (m *PinotManager) updateKnownPinotsCache(pinots []string) {
	currentPinots := make(map[string]struct{})
	for _, pinot := range pinots {
		currentPinots[pinot] = struct{}{}
	}

	// Add new pinots that are not already monitored
	for _, pinot := range pinots {
		if _, exists := m.knownPinots[pinot]; !exists {
			controller, err := m.monitorPinot(pinot)
			if err != nil {
				fmt.Printf("Unable to start monitoring %s due to error %s", pinot, err)
			}
			m.knownPinots[pinot] = controller
		}
	}

	// Unmonitor pinots that are no longer in the discovered list
	for pinot := range m.knownPinots {
		if _, exists := currentPinots[pinot]; !exists {
			err := m.unmonitorPinot(pinot)
			if err != nil {
				fmt.Printf("Encountered error while stopping monitoring of endpoint  %s due to error %s", pinot, err)
			}
			delete(m.knownPinots, pinot)
		}
	}
}

// TODO
// Receiver gets updates of which tables exist in a Pinot and sends this message to interested parties (currently
func (m *PinotManager) tableUpdateReceiverFanOut() {
}
func (m *PinotManager) monitorPinot(endpoint string) (PinotController, error) {
	var controller PinotController
	controller.URL = endpoint

	// Add a channel for table updates for this endpoint
	tablesChan := make(chan []string)
	m.tableChannels[endpoint] = tablesChan
	// setup a tablecache to refresh tables for this pinot
	tableCache := &TableCache{}
	m.tableCaches[endpoint] = tableCache

	// setup a collectorpool to collect metrics from this pinot
	workerPool := NewCollectorWorkerPool(m.numConnectorWorkers, &controller, tablesChan)
	m.workerPools[endpoint] = workerPool
	return controller, nil
}

func (m *PinotManager) unmonitorPinot(endpoint string) error {

	// TODO
	/*
	 */
	return nil
}

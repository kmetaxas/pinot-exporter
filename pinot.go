package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type PinotController struct {
	URL string `json:"url" yaml:"url"`
}

func (c *PinotController) String() string {
	return c.URL
}

/*
Get the size of the given table name in Bytes, or error

Expects a context.Context to be passed as first parameter
*/
func (c *PinotController) GetSizeForTable(ctx context.Context, tableName string) (int, error) {
	var size int = 0
	var err error

	type TableSizeResponse struct {
		Name                          string `json:"tableName"`
		ReportedSizeInBytes           int    `json:"reportedSizeInBytes"`
		EstimatedSizeInBytes          int    `json:"estimatedSizeInBytes"`
		ReportedSizePerReplicaInBytes int    `json:"reportedSizePerReplicaInBytes"`
		//offlineSegments int
		RealtimeSegments struct {
			ReportedSizeInBytes           int `json:"reportedSizeInBytes"`
			EstimatedSizeInBytes          int `json:"estimatedSizeInBytes"`
			MissingSegments               int `json:"missingSegments"`
			ReportedSizePerReplicaInBytes int `json:"reportedSizePerReplicaInBytes"`
			//segments
		}
	}
	var pinotResponse TableSizeResponse

	url := fmt.Sprintf("%s/tables/%s/size", c.URL, tableName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("pinot client: Failed to create Request obj: %s", err)
		return size, err
	}
	req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("pinot client: failed calling pinot endpoint at %s with error: %s", c.URL, err)
		return size, err
	}

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(respBody, &pinotResponse)
	if err != nil {
		log.Fatalf("Failed unmarshaling response from Pinot: %s\n", err)
	}
	size = pinotResponse.RealtimeSegments.ReportedSizeInBytes
	return size, err
}

/*
List all tables in this cluster
*/
func (c *PinotController) ListTables(ctx context.Context) ([]string, error) {
	type PinotTablesResponse struct {
		Tables []string `json:"tables"`
	}
	var pinotResponse PinotTablesResponse
	var tables []string
	var err error

	url := fmt.Sprintf("%s/tables/", c.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("pinot client: Failed to create Request obj: %s", err)
		return tables, err
	}
	req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("pinot client: failed calling pinot endpoint at %s with error: %s", c.URL, err)
		return tables, err
	}

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(respBody, &pinotResponse)
	if err != nil {
		log.Fatalf("Failed unmarshaling response from Pinot: %s\n", err)
	}
	tables = pinotResponse.Tables
	return tables, err
}

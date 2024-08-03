package main

import (
	"context"
	"fmt"
)

func main() {
	conf, err := NewConfigFromFile("testconfig.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("conf = %+v\n", conf)

	ctx := context.Background()
	tables, err := conf.PinotController.ListTables(ctx)
	fmt.Printf("Discovered tables: %+v\n", tables)
	if err != nil {
		panic(err)
	}
	for _, table := range tables {
		size, err := conf.PinotController.GetSizeForTable(ctx, table)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Size for %s= %d", table, size)
	}

}

package main

import (
	"fmt"
)

func main() {
	fmt.Println("vim-go")
	conf := NewConfig()
	fmt.Printf("conf = %+v\n", conf)
}

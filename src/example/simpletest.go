package main

import (
	"bytes"
	"fmt"
	"log"
	"scache"
	"time"
)

func main() {
	t := scache.Cache("test")

	var buf bytes.Buffer

	logger := log.New(&buf, "scache: ", log.Lshortfile)

	t.SetLogger(logger)

	_, err := t.Add("aa", 5*time.Second, "aaaaaa")
	if err != nil {
		fmt.Println("add aa error : ", err)
	}
	_, err = t.Add("aa", 6*time.Second, "cccccc")
	if err != nil {
		fmt.Println("add aa error : ", err)
	}

	item, err := t.Value("aa")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(item.Value)

	item, err = t.Value("bb")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(*t)
	time.Sleep(10 * time.Second)

	fmt.Println(&buf)
}

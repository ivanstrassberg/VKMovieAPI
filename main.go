package main

import "log"

func main() {
	store, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	store.Init()
	server := NewAPIServer(":3000", store)
	server.Run()
}

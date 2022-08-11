package main

import (
	"github.com/umputun/metrics/api"
	"github.com/umputun/metrics/storage"
	"log"
	"time"
)

const port = ":8080"

func main() {
	db := &storage.DBAccessor{}

	svc := storage.New(db)
	svc.ActivateCleanup(time.Minute) // async, exit right away

	apiService := api.Service{
		Storage: svc,
		Port:    port,
	}

	if err := apiService.Run(); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
	}
}

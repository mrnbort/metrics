package main

import (
	"context"
	"github.com/umputun/metrics/api"
	"github.com/umputun/metrics/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

const port = ":8080"

// main is the main application function
func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := storage.NewAccessor(dbConn, "metrics-service", "metrics")
	svc := storage.New(db)
	svc.ActivateCleanup(ctx, time.Minute) // async, exit right away

	auth := api.AuthMidlwr{User: "admin", Passwd: "Lapatusik"}
	apiService := api.Service{
		Storage: svc,
		Port:    port,
		Auth:    auth,
	}

	reagg := &storage.Reaggregator{
		MongoClient: dbConn,
		DbName:      "metrics-service",
		CollName:    "metrics",
		Buckets: []storage.ReaggrBucket{
			{Interval: 30 * time.Minute, Age: 24 * time.Hour, SrcType: "1m", DstType: "30m"},
		},
	}

	reagg.Do(ctx)

	if err := apiService.Run(); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
	}
}

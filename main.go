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

var opts struct {
	Port              string        `long:"port" env:"PORT" description:"port" default:":8080"`
	MongoDbUrl        string        `long:"mngdb" env:"MNG_DB" description:"MongoDB url" default:"mongodb://localhost:27017"`
	DbName            string        `long:"dbname" env:"DB_NAME" description:"MongoDB name" default:"metrics-service"`
	CollName          string        `long:"collname" env:"COLL_NAME" description:"MongoDB collection name" default:"metrics"`
	IntForgivenessPrc float64       `long:"intforgiveprc" env:"INT_FORGIVE" description:"interval forgiveness percent" default:"0.25"`
	CleanupDur        time.Duration `long:"cleanupdur" env:"CLEANUP_DUR" description:"cleanup duration" default:"1m"`
	UserName          string        `long:"username" env:"USER_NAME" description:"user name" default:"admin"`
	UserPasswd        string        `long:"userpasswd" env:"USER_PASSWD" description:"user password" default:"Lapatusik"`

	MaxExpire      time.Duration `long:"expire" env:"MAX_EXPIRE" default:"24h" description:"max lifetime"`
	MaxPinAttempts int           `long:"pinattempts" env:"PIN_ATTEMPTS" default:"3" description:"max attempts to enter pin"`
	Dbg            bool          `long:"dbg" description:"debug mode"`
}

// main is the main application function
func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.MongoDbUrl))
	if err != nil {
		panic(err)
	}

	db := storage.NewAccessor(dbConn, opts.DbName, opts.CollName, opts.IntForgivenessPrc)
	svc := storage.New(db)
	svc.ActivateCleanup(ctx, opts.CleanupDur) // async, exit right away

	auth := api.AuthMidlwr{User: opts.UserName, Passwd: opts.UserPasswd}
	apiService := api.Service{
		Storage: svc,
		Port:    opts.Port,
		Auth:    auth,
	}

	reagg := &storage.Reaggregator{
		MongoClient: dbConn,
		DbName:      opts.DbName,
		CollName:    opts.CollName,
		Buckets: []storage.ReaggrBucket{
			{Interval: 30 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}

	reagg.Do(ctx)

	if err := apiService.Run(); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/umputun/go-flags"
	"github.com/umputun/metrics/api"
	"github.com/umputun/metrics/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"os/signal"
	"time"
)

var opts struct {
	Port              string        `long:"port" env:"PORT" description:"port" default:":8080"`
	MongoDbUri        string        `long:"mngdburi" env:"MNG_DB_URI" description:"MongoDB uri" default:"mongodb://localhost:27017"`
	DbName            string        `long:"dbname" env:"DB_NAME" description:"MongoDB name" default:"metrics-service"`
	CollName          string        `long:"collname" env:"COLL_NAME" description:"MongoDB collection name" default:"metrics"`
	IntForgivenessPrc float64       `long:"intforgiveprc" env:"INT_FORGIVE" description:"interval forgiveness percent" default:"0.25"`
	CleanupDur        time.Duration `long:"cleanupdur" env:"CLEANUP_DUR" description:"cleanup duration" default:"1m"`
	UserName          string        `long:"username" env:"USER_NAME" description:"user name" default:"admin"`
	UserPasswd        string        `long:"userpasswd" env:"USER_PASSWD" description:"user password" default:"Lapatusik"`
}

// main is the main application function
func main() {

	p := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel()

	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.MongoDbUri))
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

	activateCleanup(ctx, reagg)

	if err := apiService.Run(ctx); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
		os.Exit(1)
	}
}

func activateCleanup(ctx context.Context, reagg *storage.Reaggregator) {
	go func(ctx context.Context) {
		tk := time.NewTicker(time.Hour * 24)
		defer tk.Stop()
		for range tk.C {
			err := reagg.Do(ctx) // goroutine that runs this once a day ??
			if err != nil {
				panic(err)
			}
		}
	}(ctx)
}

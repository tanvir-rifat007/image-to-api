package main

import (
	"flag"
	"log/slog"
	"os"
)

type config struct{
	port int
	env string
	
}


type application struct {
	logger *slog.Logger
	config config
  
}


func main() {
	var cfg config
	// parsing command line flags if provided
	flag.StringVar(&cfg.env,"env","development","environment")
	flag.IntVar(&cfg.port,"port",4000,"API port")
	flag.Parse()


	// using new slog logger
	logger:= slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		logger: logger,
		config: cfg,
	}

	// starting the server

	err:= app.serve()

	if err!=nil{
		logger.Error("server error","error",err)
		os.Exit(1)
		
	}




}
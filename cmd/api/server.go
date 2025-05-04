package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve()error{
	srv:= &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog: slog.NewLogLogger(app.logger.Handler(),slog.LevelError),
	}

	  shutdownError:= make(chan error)

	    go func() {
        // Create a quit channel which carries os.Signal values.
        quit := make(chan os.Signal, 1)

        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

       
        s := <-quit


        app.logger.Info("shutting down server", "signal", s.String())

				ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)

				defer cancel()



				 shutdownError <- srv.Shutdown(ctx)


    }()


	app.logger.Info("starting server","port",app.config.port,"env",app.config.env)

	err:=srv.ListenAndServe()

	if !errors.Is(err,http.ErrServerClosed){
		return err;
	}

	// Wait for the shutdown to complete
	err=<-shutdownError

	if err!=nil{
		return  err
	}
	app.logger.Info("stopped server","addr",srv.Addr)

	return nil
}
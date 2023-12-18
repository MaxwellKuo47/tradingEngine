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

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.errorLogger.Handler(), slog.LevelError),
	}
	shutdownError := make(chan error)

	go func() {
		// graceful shutdown
		quit := make(chan os.Signal, 1)
		acceptSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
		// listen for signal
		signal.Notify(quit, acceptSignals...)
		s := <-quit

		app.infoLogger.Info("shutting down server", slog.String("signal", s.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		close(app.done) // close all consumer

		// calling server shutdown
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.infoLogger.Info("completing background tasks", slog.String("addr", srv.Addr))

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.infoLogger.Info("starting server", slog.String("addr", srv.Addr), slog.String("env", app.config.env))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.infoLogger.Info("stopped server", slog.String("addr", srv.Addr))

	return nil
}

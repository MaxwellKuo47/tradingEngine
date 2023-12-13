package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// for goroutine
func (app *application) background(usage string, fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.errorLogger.Error("error background", slog.String("usage", usage), slog.Any("msg", err))
			}
			fn()
		}()
	}()
}

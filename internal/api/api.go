package api

import (
	"fmt"
	"net/http"
	"time"
	"io"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gavinmcnair/kubedb/internal/store"
	"github.com/gavinmcnair/kubedb/internal/auth"
	"github.com/gavinmcnair/kubedb/internal/config"
	"github.com/gavinmcnair/kubedb/internal/watch"
)

func StartServer(store *store.BadgerDBStore, cfg config.Config) {
	watchManager := watch.NewWatchManager()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(auth.Middleware(cfg.Token))

	r.Get("/kv/{key}", GetHandler(store))
	r.Put("/kv/{key}", PutHandler(store, watchManager))
	r.Delete("/kv/{key}", DeleteHandler(store, watchManager))
	r.Get("/watch/{key}", WatchHandler(watchManager))

	httpPort := ":8080"
	fmt.Println("Starting server on", httpPort)
	http.ListenAndServe(httpPort, r)
}

func GetHandler(store *store.BadgerDBStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		value, err := store.Get([]byte(key))
		if err != nil {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		w.Write(value)
	}
}

func PutHandler(store *store.BadgerDBStore, watchManager *watch.WatchManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")

		value, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close() // Ensure the body is closed

		if err := store.Put([]byte(key), value); err != nil {
			http.Error(w, "Failed to store key", http.StatusInternalServerError)
			return
		}

		watchManager.Notify(key, value)
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteHandler(store *store.BadgerDBStore, watchManager *watch.WatchManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		if err := store.Delete([]byte(key)); err != nil {
			http.Error(w, "Failed to delete key", http.StatusInternalServerError)
			return
		}

		watchManager.Notify(key, nil) // Notify with nil to indicate deletion
		w.WriteHeader(http.StatusNoContent)
	}
}

func WatchHandler(watchManager *watch.WatchManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		ctx := r.Context()

		// Register for notifications
		ch := watchManager.Watch(key)

		// Listen for changes in a loop
		for {
			select {
			case value := <-ch:
				w.Write(value)
				return
			case <-time.After(30 * time.Second): // Timeout every 30 seconds
				http.Error(w, "No change detected", http.StatusNoContent)
				watchManager.CancelWatch(key, ch)
				return
			case <-ctx.Done(): // Client disconnected
				watchManager.CancelWatch(key, ch)
				return
			}
		}
	}
}


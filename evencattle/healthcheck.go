package evencattle

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	router          = mux.NewRouter()
	healthCheckPort = ":9777"
)

func healthCheckResponse(w http.ResponseWriter, req *http.Request) {
	// responde to the ping!
	w.Write([]byte("pong"))
}

func serve() {
	go log.Fatal(http.ListenAndServe(healthCheckPort, router))
}

func StartHealthCheck() {
	router.HandleFunc("/ping", healthCheckResponse).Methods("GET", "HEAD").Name("healthcheck")
	log.Info("healthcheck handler is listening on ", healthCheckPort)
	go serve()
}

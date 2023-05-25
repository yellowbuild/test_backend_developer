package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

const INTERFACE = "127.0.0.1"
const PORT = "8000"
const SERVER = INTERFACE + ":" + PORT

func showSpots(w http.ResponseWriter, r *http.Request) {
	latitude := r.URL.Query().Get("latitude")
	longitude := r.URL.Query().Get("latitude")
	radius_meters := r.URL.Query().Get("radius")
	radius_type := r.URL.Query().Get("type")

	if latitude != "" && longitude != "" {
		if _, err := strconv.ParseFloat(radius_meters, 64); err == nil {
			if radius_type == "square" || radius_type == "circle" {
				fmt.Fprintf(w, "%s", string("Success"))
			} else {
				http.Error(w, "Invalid parameter: type", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Invalid parameter: radius", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Invalid parameter: latitude||longitude", http.StatusBadRequest)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", showSpots).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(r)
	http.ListenAndServe(SERVER, n)
}

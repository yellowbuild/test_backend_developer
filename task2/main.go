package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const INTERFACE = "127.0.0.1"
const PORT = "8000"
const SERVER = INTERFACE + ":" + PORT

type Spot struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Website     sql.NullString `json:"website"`
	Coordinates string         `json:"coordinates"`
	Description sql.NullString `json:"description"`
	Rating      float64        `json:"rating"`
	Distance    float64        `json:"distance,omitempty"`
}

type SpotsResponse struct {
	Spots []Spot `json:"spots"`
}

const (
	DbDriver   = "postgres"
	DbUser     = "postgres"
	DbPassword = "tryhackme"
	DbName     = "spots"
)

func main() {
	http.HandleFunc("/spots", handleSpotsRequest)
	log.Fatal(http.ListenAndServe(SERVER, nil))
}

func handleSpotsRequest(w http.ResponseWriter, r *http.Request) {
	latitude, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	radius, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	if err != nil {
		http.Error(w, "Invalid radius", http.StatusBadRequest)
		return
	}

	locationType := r.URL.Query().Get("type")
	if locationType != "circle" && locationType != "square" {
		http.Error(w, "Invalid location type. Valid types are 'circle' and 'square'", http.StatusBadRequest)
		return
	}

	spots, err := getSpotsInArea(latitude, longitude, radius, locationType)
	if err != nil {
		http.Error(w, "Error retrieving spots", http.StatusInternalServerError)
		return
	}

	response := SpotsResponse{Spots: spots}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func getSpotsInArea(latitude, longitude, radius float64, locationType string) ([]Spot, error) {
	db, err := sql.Open(DbDriver, fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", DbUser, DbPassword, DbName))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer db.Close()

	query := ""
	switch locationType {
	case "circle":
		query = fmt.Sprintf(`
			SELECT
				id, name, website, ST_AsText(coordinates), description, rating,
				ST_Distance(coordinates::geography, 'SRID=4326;POINT(%f %f)') AS distance
			FROM
				"MY_TABLE"
			WHERE
				ST_DWithin(coordinates::geography, 'SRID=4326;POINT(%f %f)', %f)
			ORDER BY
				distance, rating DESC;
		`, longitude, latitude, longitude, latitude, radius)
	case "square":
		// Calculate the latitude and longitude boundaries for the square area
		latBoundary := 180 * radius / (math.Pi * 6371)                // Approximation for latitude
		lngBoundary := latBoundary / math.Cos(latitude*math.Pi/180.0) // Adjusted for longitude

		query = fmt.Sprintf(`
			SELECT
				id, name, website, ST_AsText(coordinates), description, rating
			FROM
				"MY_TABLE"
			WHERE
				ST_X(coordinates::geometry) >= %f - %f AND ST_X(coordinates::geometry) <= %f + %f AND
				ST_Y(coordinates::geometry) >= %f - %f AND ST_Y(coordinates::geometry) <= %f + %f
			ORDER BY
				ST_Distance(coordinates::geography, 'SRID=4326;POINT(%f %f)')
		`, latitude, latBoundary, latitude, latBoundary, longitude, lngBoundary, longitude, lngBoundary, latitude, longitude)
	}

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	spots := []Spot{}
	for rows.Next() {
		var spot Spot
		err := rows.Scan(&spot.ID, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating, &spot.Distance)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		spots = append(spots, spot)
	}

	return spots, nil
}

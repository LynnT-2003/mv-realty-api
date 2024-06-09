package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
    "github.com/gorilla/handlers"
	"github.com/joho/godotenv"
)

var apiKey string

type Condo struct {
	// facilities should be arrays of strings
	CondoId     int          `json:"condoId"`
	CondoName   string       `json:"condoName"`
	Address     string       `json:"address"`
	City        string       `json:"city"`
	Facilities  string       `json:"facilities"`
	Description string       `json:"description"`
	Types       []TypeOfRoom `json:"types"`
}

type TypeOfRoom struct {
	TypeId      string `json:"typeId"`
	TypeName    string `json:"typeName"`
	Description string `json:"description"`
}

type Listing struct {
	ListingId   int    `json:"listingId"`
	CondoId     int    `json:"condoId"`
	TypeId      string `json:"typeId"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

var (
	condos   []Condo
	listings []Listing
)

func apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key != apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getCondos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(condos)
}

func getListings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

func getListingsByCondoId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	condoId, err := strconv.Atoi(params["condoId"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}

	var condoListings []Listing
	for _, listing := range listings {
		if listing.CondoId == condoId {
			condoListings = append(condoListings, listing)
		}
	}

	if len(condoListings) == 0 {
		http.Error(w, "No listings found for the specified Condo ID", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(condoListings)
}

func getListingsByStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	status, ok := params["status"]
	if !ok {
		http.Error(w, "Invalid status parameter", http.StatusBadRequest)
		return
	}

	var condoListings []Listing
	for _, listing := range listings {
		if listing.Status == status {
			condoListings = append(condoListings, listing)
		}
	}
	json.NewEncoder(w).Encode(condoListings)
}

func getListingsDetailFromListingsByCondoId(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	condoId, err := strconv.Atoi(params["condoId"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}
	listingId, err := strconv.Atoi(params["listingId"])
	if err != nil {
		http.Error(w, "Invalid Listing ID", http.StatusBadRequest)
		return
	}

	for _, listing := range listings {
		if listing.CondoId == condoId && listing.ListingId == listingId {
			json.NewEncoder(w).Encode(listing)
			break
		}
	}
}

func getListingsByCondoIdAndStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	condoId, err := strconv.Atoi(params["condoId"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}
	status, ok := params["status"]
	if !ok {
		http.Error(w, "Invalid status parameter", http.StatusBadRequest)
		return
	}

	var condoListings []Listing
	for _, listing := range listings {
		if listing.CondoId == condoId && listing.Status == status {
			condoListings = append(condoListings, listing)
		}
	}

	if len(condoListings) == 0 {
		http.Error(w, "No listings found for the specified Status in Condo Listings.", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(condoListings)
}

func getListingsByCondoIdAndType(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	condoId, err := strconv.Atoi(params["condoId"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}
	typeId, ok := params["type"]
	if !ok {
		http.Error(w, "Invalid type parameter", http.StatusBadRequest)
		return
	}

	var condoListings []Listing
	for _, listing := range listings {
		if listing.CondoId == condoId && listing.TypeId == typeId {
			condoListings = append(condoListings, listing)
		}
	}

	if len(condoListings) == 0 {
		http.Error(w, "No listings found for the specified Type in Condo Listings.", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(condoListings)
}

func createListing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newListing Listing
	if err := json.NewDecoder(r.Body).Decode(&newListing); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate CondoId
	var validCondo *Condo
	for _, condo := range condos {
		if condo.CondoId == newListing.CondoId {
			validCondo = &condo
			break
		}
	}

	if validCondo == nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}

	// Validate TypeId
	validTypeId := false
	for _, roomType := range validCondo.Types {
		if roomType.TypeId == newListing.TypeId {
			validTypeId = true
			break
		}
	}

	if !validTypeId {
		http.Error(w, "Invalid Type ID", http.StatusBadRequest)
		return
	}

	// Check for duplicate ListingId
	for _, listing := range listings {
		if listing.ListingId == newListing.ListingId {
			http.Error(w, "Listing ID already exists", http.StatusBadRequest)
			return
		}
	}

	// Add the new listing
	listings = append(listings, newListing)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newListing)
}

func deleteListing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	listingIdToDelete, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}

	for index, item := range listings {
		if item.ListingId == listingIdToDelete {
			listings = append(listings[:index], listings[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(listings)
}

func createCondo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newCondo Condo
	if err := json.NewDecoder(r.Body).Decode(&newCondo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check for duplicate CondoId and CondoName
	for _, condo := range condos {
		if condo.CondoId == newCondo.CondoId || condo.CondoName == newCondo.CondoName {
			http.Error(w, "Condo already exists", http.StatusBadRequest)
			return
		}
	}

	condos = append(condos, newCondo)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCondo)
}

func deleteCondo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	condoIdToDelete, err := strconv.Atoi(params["condoId"])
	if err != nil {
		http.Error(w, "Invalid Condo ID", http.StatusBadRequest)
		return
	}

	for index, item := range condos {
		if item.CondoId == condoIdToDelete {
			condos = append(condos[:index], condos[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(condos)
}

func init() {
	// Initialize with sample data
	condos = []Condo{
		{
			CondoId:     1,
			CondoName:   "Sunset Plaza",
			Address:     "123 Sunshine Blvd",
			City:        "Sunnyville",
			Facilities:  "Pool, Gym, Parking",
			Description: "A luxurious condo with all amenities.",
			Types: []TypeOfRoom{
				{TypeId: "SP1", TypeName: "1 Bedroom", Description: "One bedroom condo."},
				{TypeId: "SP2", TypeName: "2 Bedroom", Description: "Two bedroom condo."},
			},
		},
		{
			CondoId:     2,
			CondoName:   "Ocean Breeze",
			Address:     "456 Ocean View",
			City:        "Beach City",
			Facilities:  "Pool, Sauna, Parking",
			Description: "Condo with stunning ocean views.",
			Types: []TypeOfRoom{
				{TypeId: "OB1-ov", TypeName: "1 Bedroom", Description: "One bedroom condo with ocean view."},
				{TypeId: "OB2-ov", TypeName: "2 Bedroom", Description: "Two bedroom condo with ocean view."},
			},
		},
	}

	listings = []Listing{
		{
			ListingId:   1,
			CondoId:     1,
			TypeId:      "SP1",
			Price:       300000,
			Description: "Beautiful one bedroom condo in Sunset Plaza.",
			Status:      "available-for-sale",
		},
		{
			ListingId:   2,
			CondoId:     1,
			TypeId:      "SP2",
			Price:       450000,
			Description: "Spacious two bedroom condo in Sunset Plaza.",
			Status:      "available-for-rent",
		},
		{
			ListingId:   3,
			CondoId:     2,
			TypeId:      "OB1-ov",
			Price:       350000,
			Description: "Cozy one bedroom condo in Ocean Breeze.",
			Status:      "available-for-sale",
		},
		{
			ListingId:   4,
			CondoId:     2,
			TypeId:      "OB2-ov",
			Price:       500000,
			Description: "Luxurious two bedroom condo in Ocean Breeze.",
			Status:      "available-for-rent",
		},
	}
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file:", err)
    }

    apiKey = os.Getenv("API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable not set")
    }

    // Initialize Router
    r := mux.NewRouter()

    // API Key middleware applied to all routes
    api := r.PathPrefix("/api").Subrouter()
    api.Use(apiKeyMiddleware)

    // GET requests
    // TODO: getCondoById and getListingById
    r.HandleFunc("/condos", getCondos).Methods("GET")
    r.HandleFunc("/listings", getListings).Methods("GET")
    r.HandleFunc("/condos/{condoId}/listings", getListingsByCondoId).Methods("GET")
    r.HandleFunc("/condos/{condoId}/type/{type}", getListingsByCondoIdAndType).Methods("GET")
    r.HandleFunc("/listings/status/{status}", getListingsByStatus).Methods("GET")
    r.HandleFunc("/condos/{condoId}/listings/status/{status}", getListingsByCondoIdAndStatus).Methods("GET")
    r.HandleFunc("/condos/{condoId}/listings/{listingId}", getListingsDetailFromListingsByCondoId).Methods("GET")

    // POST request
    r.HandleFunc("/listings", createListing).Methods("POST")
    r.HandleFunc("/condos", createCondo).Methods("POST")

    // TODO: PUT requests

    // DELETE request
    r.HandleFunc("/listings/{id}", deleteListing).Methods("DELETE")
    r.HandleFunc("/condos/{condoId}", deleteCondo).Methods("DELETE")
    // TODO: When deleting a Condo, delete all listings related to the Condo

    // Enable CORS
    cors := handlers.CORS(
        handlers.AllowedOrigins([]string{"*"}),    // Allow requests from all origins
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        handlers.AllowedHeaders([]string{"Content-Type", "X-API-Key"}),
    )

    // Create a new handler with CORS middleware
    handler := cors(r)

    // Ports and Server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000" // Default port to 8000 if PORT environment variable is not set
    }
    fmt.Println("Server is running on port:", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}

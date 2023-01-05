package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
)

// Define Receipt Structure
type Receipt struct {
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items []Item `json:"items"`
	Total string `json:"total"`
}

// Define Item Structure
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price string `json:"price"`
}

// Declare a global Receipts array to populate in main function to simulate a database
var Receipts []Receipt

// POST method that processes receipts
func processReceipts(w http.ResponseWriter, r *http.Request){
	reqBody, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(reqBody))
	// fmt.Fprintf(w, "%+v", string(reqBody))
	fmt.Println("Endpoint Hit: POST Process Receipts")
	var receipt Receipt
	json.Unmarshal(reqBody, &receipt) 
	Receipts = append(Receipts, receipt)
	json.NewEncoder(w).Encode(receipt)
}

// GET method that returns total points
func getPoints(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "GET: Get Points")
}

// Helper function to return all receipts in database to verify that processed receipt is added to Receipts
func returnAllReceipts(w http.ResponseWriter, r *http.Request){
	fmt.Println("Endpoint Hit: returnAllReceipts")
	json.NewEncoder(w).Encode(Receipts)
}

func handleRequests() {
	http.HandleFunc("/receipts/process/", processReceipts)
	http.HandleFunc("/receipts/{id}/points/", getPoints)
	http.HandleFunc("/receipts/", returnAllReceipts)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func main() {
	handleRequests()
}
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Receipt Structure Definition
type Receipt struct {
	Id string `json:"id"`
	Retailer string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items []Item `json:"items"`
	Total string `json:"total"`
}

// Item Structure Definition
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price string `json:"price"`
}

// Global Receipts map to simulate a database, with id mapped to Receipt object
var (
	Receipts = make(map[string]Receipt)
)

// POST method that processes receipts
func processReceipts(w http.ResponseWriter, r *http.Request){
	reqBody, _ := ioutil.ReadAll(r.Body)
	var receipt Receipt
	json.Unmarshal(reqBody, &receipt)
	// Validate input receipt JSON is in valid format
	isValid, errResponse := validateReceiptInput(receipt)
	if (!isValid) {
		json.NewEncoder(w).Encode(errResponse)
		return
	}
	// Assign unique ID to receipt
	receipt.Id = uuid.New().String()
	// Add receipt to Receipts database
	Receipts[receipt.Id] = receipt
	// Build and return response
	response := map[string]string{"id":receipt.Id}
	json.NewEncoder(w).Encode(response)
}

// GET method that returns total points for a given receipt
func getPoints(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	id := vars["id"]
	// Get receipt with specified id
	receipt := getReceipt(id)
	// Validate id 
	if receipt == nil {
		// Build and return error response
		response := map[string]string{"error":"No receipt found. Invalid id: " + id}
		json.NewEncoder(w).Encode(response)
		return
	}
	// Process receipt and calculate total points
	var points int = 0
	// Check retailer name
	for _, char := range receipt.Retailer {
		if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || ('0' <= char && char <= '9') {
			points++
		}
	}
	// Check total
	totalCentsPlace := string(receipt.Total[len(receipt.Total)-2:])
	// Check if total is a whole number
	if (totalCentsPlace == "00") {
		points += 50
	}
	// Check if total is a multiple of 0.25
	if totalCentsPlace == "00" || totalCentsPlace == "25" || totalCentsPlace == "50" || totalCentsPlace == "75" {
		points += 25
	}
	// Add 5 points for every 2 items on the receipt
	numItems := len(receipt.Items)
	points += 5*(numItems / 2)
	// Check trimmed length for each item description
	for _, item := range receipt.Items {
		// Trim off beginning and trailing white space
		descriptionLength := len(strings.TrimSpace(item.ShortDescription))
		if (descriptionLength % 3 == 0) {
			itemPrice, err := strconv.ParseFloat(item.Price, 64)
			if err == nil {
				points += int(math.Ceil(itemPrice * 0.2))
			}
		}
	}
	// Check if purchaseDate is odd
	dateString := string(receipt.PurchaseDate[len(receipt.PurchaseDate)-2:])
	date, _ := strconv.ParseFloat(dateString, 64)
	if int(date) % 2 != 0 {
		points += 6
	}
	// Check if purchaseTime is between 2-4pm (i.e. 14:00 - 16:00)
	purchaseTime, _ := strconv.ParseFloat(receipt.PurchaseTime[:2] + "." + receipt.PurchaseTime[3:], 64)
	if purchaseTime > 14.00 && purchaseTime < 16.00 {
		points += 10
	}
	// Build and return response
	response := map[string]int{"points":points}
	json.NewEncoder(w).Encode(response)
}

// Helper function to validate receipt input
func validateReceiptInput(receipt Receipt)(bool, map[string]string) {
	errorMap := make(map[string]string)
	// Verify purchase date is in YYYY-MM-DD format
	_, errYear := strconv.Atoi(string(receipt.PurchaseDate[:4]));
	_, errMonth := strconv.Atoi(string(receipt.PurchaseDate[5:7]))
	_, errDay := strconv.Atoi(string(receipt.PurchaseDate[8:]))
	if (len(receipt.PurchaseDate) != 10 || 
			string(receipt.PurchaseDate[4]) != "-" || 
			string(receipt.PurchaseDate[7]) != "-" ||
			errYear != nil ||
			errMonth != nil ||
			errDay != nil) {
		errorMap["purchaseDate"] = "Invalid input: " + receipt.PurchaseDate
	}
	// Verify purchase time is in hh:mm format
	_, errHour := strconv.Atoi(string(receipt.PurchaseTime[0:2]))
	_, errMin := strconv.Atoi(string(receipt.PurchaseTime[3:]))
	if (len(receipt.PurchaseTime) != 5 || 
			string(receipt.PurchaseTime[2]) != ":" || 
			errHour != nil || 
			errMin != nil) {
		errorMap["purchaseTime"] = "Invalid input: " + receipt.PurchaseTime
	}
	// Verify total is a float and ends in .xx
	_, err := strconv.ParseFloat(receipt.Total, 64)
	if (err != nil || (string(receipt.Total[len(receipt.Total)-3]) != ".")) {
		errorMap["total"] = "Invalid input: " + receipt.Total
	}

	if (len(errorMap) != 0) {
		errorMap["error"] = "Receipt JSON contains invalid inputs"
		return false, errorMap
	}
	// No errors found
	return true, errorMap
}

// Helper function to get receipt with specified id
func getReceipt(id string)(*Receipt){
	receipt, exists := Receipts[id]
	if (exists == true) {
		return &receipt
	}
	// No receipt found
	return nil
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/receipts/process/", processReceipts).Methods("POST")
	myRouter.HandleFunc("/receipts/{id}/points/", getPoints)
	log.Fatal(http.ListenAndServe(":5000", myRouter))
}

func main() {
	handleRequests()
}
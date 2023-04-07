package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ExchangeRequest struct {
	Amount float64 `json:"amount"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Rate   float64 `json:"rate"`
}

type ExchangeResponse struct {
	ValorConvertido float64 `json:"valorConvertido"`
	Simbolo         string  `json:"simboloMoeda"`
}

var coins = map[string]string{
	"USD": "$",
	"BRL": "R$",
	"EUR": "€",
	"BTC": "₿",
}

func exchangeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := validateParams(vars)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := obj.Amount * obj.Rate
	simbol := coins[vars["to"]]

	response := ExchangeResponse{
		ValorConvertido: result,
		Simbolo:         simbol,
	}

	db, err := sql.Open("mysql", "admin:password@tcp(127.0.0.1:3306)/bancogenesis")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO exchange(amount, from_currency, to_currency, rate, converted_value, symbol) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(obj.Amount, obj.From, obj.To, obj.Rate, response.ValorConvertido, response.Simbolo)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func validateParams(vars map[string]string) (*ExchangeRequest, error) {

	amount, err := strconv.ParseFloat(vars["amount"], 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid amount: %v", err)
	}

	rate, err2 := strconv.ParseFloat(vars["rate"], 64)
	if err2 != nil {
		return nil, fmt.Errorf("Invalid rate: %v", err)
	}

	from := vars["from"]
	if _, ok := coins[from]; !ok {
		return nil, fmt.Errorf("Invalid from currency: %v", from)
	}

	to := vars["to"]
	if _, ok := coins[to]; !ok {
		return nil, fmt.Errorf("Invalid to currency: %v", to)
	}

	if to == from {
		return nil, fmt.Errorf("From/to is same country: %v", to)
	}

	exchange := ExchangeRequest{
		Amount: amount,
		From:   from,
		To:     to,
		Rate:   rate,
	}

	return &exchange, nil

}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/exchange/{amount}/{from}/{to}/{rate}", exchangeHandler).Methods("POST")

	fmt.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

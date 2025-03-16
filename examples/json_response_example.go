package main

import (
    "fmt"
    "net/http"
    "github\.com/JackovAlltrades/go-toolbox"
)

type Person struct {
    Name    string json:"name"
    Age     int    json:"age"
    Email   string json:"email"
}

func main() {
    http.HandleFunc("/json", jsonHandler)
    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", nil)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    tools := &toolbox.Tools{}
    
    person := Person{
        Name:  "John Doe",
        Age:   30,
        Email: "john@example.com",
    }
    
    err := tools.WriteJSON(w, http.StatusOK, person, nil)
    if err != nil {
        http.Error(w, "Error writing JSON", http.StatusInternalServerError)
    }
}


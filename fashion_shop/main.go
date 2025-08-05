package main

import (
    "fmt"
    "html/template"
    "net/http"
    "strconv"
)

type Order struct {
    OrderID    string
    Contact    string
    Size       string
    Quantity   int
    TotalPrice float64
}

var orders []Order
var lastOrderNumber = 0

var priceMap = map[string]float64{
    "XS": 600,
    "S":  800,
    "M":  900,
    "L":  1000,
    "XL": 1100,
    "XXL": 1200,
}

func generateOrderID() string {
    lastOrderNumber++
    return fmt.Sprintf("ODR#%05d", lastOrderNumber)
}

func placeOrderPage(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        t, _ := template.ParseFiles("templates/form.html")
        t.Execute(w, nil)
    } else if r.Method == http.MethodPost {
        contact := r.FormValue("contact")
        size := r.FormValue("size")
        qtyStr := r.FormValue("qty")
        qty, _ := strconv.Atoi(qtyStr)
        total := priceMap[size] * float64(qty)

        order := Order{
            OrderID:    generateOrderID(),
            Contact:    contact,
            Size:       size,
            Quantity:   qty,
            TotalPrice: total,
        }

        orders = append(orders, order)

        t, _ := template.ParseFiles("templates/success.html")
        t.Execute(w, order)
    }
}

func homePage(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/place-order", http.StatusSeeOther)
}

func main() {
    http.HandleFunc("/", homePage)
    http.HandleFunc("/place-order", placeOrderPage)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    fmt.Println("Server started at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

// Web version with features 2 to 6 added
package main

import (
    "fmt"
    "html/template"
    "net/http"
    "strconv"
)

type Order struct {
    OrderID     string
    CustomerID  string
    Size        string
    Quantity    int
    TotalAmount float64
    Status      string
}

var orders []Order
var lastOrderNumber = 0
var priceMap = map[string]float64{
    "XS": 600, "S": 800, "M": 900, "L": 1000, "XL": 1100, "XXL": 1200,
}
var statuses = []string{"PROCESSING", "DELIVERING", "DELIVERED"}

func generateOrderID() string {
    lastOrderNumber++
    return fmt.Sprintf("ODR#%05d", lastOrderNumber)
}

func home(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/home.html"))
    tmpl.Execute(w, nil)
}


func placeOrderPage(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        tmpl := template.Must(template.ParseFiles("templates/form.html"))
        tmpl.Execute(w, nil)
    } else if r.Method == http.MethodPost {
        contact := r.FormValue("contact")
        size := r.FormValue("size")
        qty, _ := strconv.Atoi(r.FormValue("qty"))
        amount := priceMap[size] * float64(qty)
        order := Order{
            OrderID:     generateOrderID(),
            CustomerID:  contact,
            Size:        size,
            Quantity:    qty,
            TotalAmount: amount,
            Status:      statuses[0],
        }
        orders = append(orders, order)
        tmpl := template.Must(template.ParseFiles("templates/success.html"))
        tmpl.Execute(w, order)
    }
}

func searchCustomer(w http.ResponseWriter, r *http.Request) {
    contact := r.URL.Query().Get("contact")
    var found []Order
    for _, o := range orders {
        if o.CustomerID == contact {
            found = append(found, o)
        }
    }
    tmpl := template.Must(template.ParseFiles("templates/search_customer.html"))
    tmpl.Execute(w, found)
}

func searchOrder(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("id")
    for _, o := range orders {
        if o.OrderID == orderID {
            tmpl := template.Must(template.ParseFiles("templates/search_order.html"))
            tmpl.Execute(w, o)
            return
        }
    }
    fmt.Fprint(w, "Order not found")
}

func viewReports(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/reports.html"))
    tmpl.Execute(w, orders)
}

func changeStatus(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    for i, o := range orders {
        if o.OrderID == id && o.Status != "DELIVERED" {
            if o.Status == "PROCESSING" {
                orders[i].Status = "DELIVERING"
            } else if o.Status == "DELIVERING" {
                orders[i].Status = "DELIVERED"
            }
            fmt.Fprint(w, "Status updated")
            return
        }
    }
    fmt.Fprint(w, "Invalid Order ID or already delivered")
}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    for i, o := range orders {
        if o.OrderID == id {
            orders = append(orders[:i], orders[i+1:]...)
            fmt.Fprint(w, "Order deleted")
            return
        }
    }
    fmt.Fprint(w, "Order not found")
}

func main() {
    http.HandleFunc("/", home) 
    http.HandleFunc("/form", placeOrderPage)
    http.HandleFunc("/search-customer", searchCustomer)
    http.HandleFunc("/search-order", searchOrder)
    http.HandleFunc("/reports", viewReports)
    http.HandleFunc("/change-status", changeStatus)
    http.HandleFunc("/delete-order", deleteOrder)
    fmt.Println("Server running at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}




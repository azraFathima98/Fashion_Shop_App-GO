package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
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
var db *sql.DB
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

func searchCustomerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/search_customer_form.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		contact := r.FormValue("contact")
		var found []Order
		for _, o := range orders {
			if o.CustomerID == contact {
				found = append(found, o)
			}
		}
		tmpl := template.Must(template.ParseFiles("templates/search_customer_results.html"))
		tmpl.Execute(w, found)
	}
}

func searchOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/search_order_form.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		orderID := strings.TrimSpace(r.FormValue("orderid"))
		for _, o := range orders {
			if o.OrderID == orderID {
				tmpl := template.Must(template.ParseFiles("templates/search_order_results.html"))
				tmpl.Execute(w, o)
				return
			}
		}
		tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
		tmpl.Execute(w, nil)
	}
}

type ReportData struct {
    Orders      []Order
    TotalOrders int
    TotalAmount float64
}

func viewReports(w http.ResponseWriter, r *http.Request) {
    total := 0.0
    for _, o := range orders {
        total += o.TotalAmount
    }
    data := ReportData{
        Orders:      orders,
        TotalOrders: len(orders),
        TotalAmount: total,
    }
    tmpl := template.Must(template.ParseFiles("templates/reports.html"))
    tmpl.Execute(w, data)
}


func changeStatusPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/change_status_form.html"))
		tmpl.Execute(w, orders)
	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")
		for i, o := range orders {
			if o.OrderID == id && o.Status != "DELIVERED" {
				if o.Status == "PROCESSING" {
					orders[i].Status = "DELIVERING"
				} else if o.Status == "DELIVERING" {
					orders[i].Status = "DELIVERED"
				}
				tmpl := template.Must(template.ParseFiles("templates/status_updated.html"))
				tmpl.Execute(w, orders[i])
				return
			}
		}
		tmpl := template.Must(template.ParseFiles("templates/status_error.html"))
		tmpl.Execute(w, nil)
	}
}

func deleteOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/delete_order_form.html"))
		tmpl.Execute(w, orders)
	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")
		for i, o := range orders {
			if o.OrderID == id {
				deletedOrder := orders[i]
				orders = append(orders[:i], orders[i+1:]...)
				tmpl := template.Must(template.ParseFiles("templates/order_deleted.html"))
				tmpl.Execute(w, deletedOrder)
				return
			}
		}
		tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
		tmpl.Execute(w, nil)
	}
}

func main() {
	var err error
    db, err = sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/fashion_shop")
    if err != nil {
        panic(err)
    }
    defer db.Close()
	http.HandleFunc("/", home)
	http.HandleFunc("/place-order", placeOrderPage)
	http.HandleFunc("/search-customer", searchCustomerPage)
	http.HandleFunc("/search-order", searchOrderPage)
	http.HandleFunc("/reports", viewReports)
	http.HandleFunc("/change-status", changeStatusPage)
	http.HandleFunc("/delete-order", deleteOrderPage)
	
	fmt.Println("Server running at http://localhost:8080")
	
	http.ListenAndServe(":8080", nil)
}
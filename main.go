package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

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
		// Show the search form
		tmpl := template.Must(template.ParseFiles("templates/search_order_form.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		orderID := strings.TrimSpace(r.FormValue("orderid"))
		if orderID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		var foundOrder *Order
		for _, o := range orders {
			if o.OrderID == orderID {
				foundOrder = &o
				break
			}
		}

		if foundOrder == nil {
			
			tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
			tmpl.Execute(w, nil)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/search_order_results.html"))
		tmpl.Execute(w, foundOrder)
	}
}



type ReportData struct {
    Orders      []Order
    TotalOrders int
    TotalAmount float64
}

func viewReports(w http.ResponseWriter, r *http.Request) {
	var totalRevenue float64
	for _, o := range orders {
		totalRevenue += o.TotalAmount
	}

	data := ReportData{
		Orders:      orders,
		TotalOrders: len(orders),
		TotalAmount: totalRevenue,
	}

	tmpl := template.Must(template.ParseFiles("templates/reports.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}




func changeStatusPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		
		tmpl := template.Must(template.ParseFiles("templates/change_status_form.html"))
		tmpl.Execute(w, orders)
	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")

		
		var orderIndex = -1
		for i, o := range orders {
			if o.OrderID == id {
				orderIndex = i
				break
			}
		}

		if orderIndex == -1 {
			
			tmpl := template.Must(template.ParseFiles("templates/status_error.html"))
			tmpl.Execute(w, nil)
			return
		}

		
		currentStatus := orders[orderIndex].Status
		var newStatus string

		if currentStatus == "PROCESSING" {
			newStatus = "DELIVERING"
		} else if currentStatus == "DELIVERING" {
			newStatus = "DELIVERED"
		} else {
			
			tmpl := template.Must(template.ParseFiles("templates/status_error.html"))
			tmpl.Execute(w, nil)
			return
		}

		
		orders[orderIndex].Status = newStatus

		
		tmpl := template.Must(template.ParseFiles("templates/status_updated.html"))
		tmpl.Execute(w, orders[orderIndex])
	}
}


func deleteOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		
		tmpl := template.Must(template.ParseFiles("templates/delete_order_form.html"))
		tmpl.Execute(w, orders)
	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")

		
		index := -1
		for i, o := range orders {
			if o.OrderID == id {
				index = i
				break
			}
		}

		if index == -1 {
			
			tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
			tmpl.Execute(w, nil)
			return
		}

		
		orders = append(orders[:index], orders[index+1:]...)

		
		tmpl := template.Must(template.ParseFiles("templates/order_deleted.html"))
		tmpl.Execute(w, struct{ OrderID string }{OrderID: id})
	}
}


func main() {
	

	http.HandleFunc("/", home)
	http.HandleFunc("/place-order", placeOrderPage)
	http.HandleFunc("/search-customer", searchCustomerPage) 
	http.HandleFunc("/search-order", searchOrderPage)
	http.HandleFunc("/reports", viewReports)
	http.HandleFunc("/change-status", changeStatusPage)
	http.HandleFunc("/delete-order", deleteOrderPage)

	fmt.Println("Server running at http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
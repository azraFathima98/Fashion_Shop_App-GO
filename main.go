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
		_, err := db.Exec("INSERT INTO orders (order_id, customer_id, size, quantity, total_amount, status) VALUES (?, ?, ?, ?, ?, ?)",
    	order.OrderID, order.CustomerID, order.Size, order.Quantity, order.TotalAmount, order.Status)
		if err != nil {
    	http.Error(w, "Database error", http.StatusInternalServerError)
    	return
}
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
		rows, err := db.Query("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders WHERE customer_id = ?", contact)
	if err != nil {
	http.Error(w, "Database error", http.StatusInternalServerError)
	return
}
	defer rows.Close()

	var found []Order
	for rows.Next() {
	var o Order
	rows.Scan(&o.OrderID, &o.CustomerID, &o.Size, &o.Quantity, &o.TotalAmount, &o.Status)
	found = append(found, o)

}

		tmpl := template.Must(template.ParseFiles("templates/search_customer_results.html"))
		tmpl.Execute(w, found)
	}
}

func searchOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// show the form
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

		var o Order
		row := db.QueryRow("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders WHERE order_id = ?", orderID)
		err := row.Scan(&o.OrderID, &o.CustomerID, &o.Size, &o.Quantity, &o.TotalAmount, &o.Status)
		if err == sql.ErrNoRows {
			// Order not found
			tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
			tmpl.Execute(w, nil)
			return
		} else if err != nil {
			// DB error
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Order found, display it
		tmpl := template.Must(template.ParseFiles("templates/search_order_results.html"))
		tmpl.Execute(w, o)
	}
}



type ReportData struct {
    Orders      []Order
    TotalOrders int
    TotalAmount float64
}

func viewReports(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	var totalRevenue float64

	for rows.Next() {
		var o Order
		err := rows.Scan(&o.OrderID, &o.CustomerID, &o.Size, &o.Quantity, &o.TotalAmount, &o.Status)
		if err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
		totalRevenue += o.TotalAmount
	}

	data := ReportData{
		Orders:      orders,
		TotalOrders: len(orders),
		TotalAmount: totalRevenue,
	}

	tmpl := template.Must(template.ParseFiles("templates/reports.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}




func changeStatusPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders")
if err != nil {
	http.Error(w, "Database error", http.StatusInternalServerError)
	return
}
defer rows.Close()

var allOrders []Order
for rows.Next() {
	var o Order
	rows.Scan(&o.OrderID, &o.CustomerID, &o.Size, &o.Quantity, &o.TotalAmount, &o.Status)
	allOrders = append(allOrders, o)
}
tmpl := template.Must(template.ParseFiles("templates/change_status_form.html"))
tmpl.Execute(w, allOrders)

	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")

var status string
err := db.QueryRow("SELECT status FROM orders WHERE order_id = ?", id).Scan(&status)
if err == sql.ErrNoRows {
	tmpl := template.Must(template.ParseFiles("templates/status_error.html"))
	tmpl.Execute(w, nil)
	return
} else if err != nil {
	http.Error(w, "Database error", http.StatusInternalServerError)
	return
}

newStatus := ""
if status == "PROCESSING" {
	newStatus = "DELIVERING"
} else if status == "DELIVERING" {
	newStatus = "DELIVERED"
} else {
	tmpl := template.Must(template.ParseFiles("templates/status_error.html"))
	tmpl.Execute(w, nil)
	return
}

_, err = db.Exec("UPDATE orders SET status = ? WHERE order_id = ?", newStatus, id)
if err != nil {
	http.Error(w, "Database update error", http.StatusInternalServerError)
	return
}

row := db.QueryRow("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders WHERE order_id = ?", id)
var updated Order
row.Scan(&updated.OrderID, &updated.CustomerID, &updated.Size, &updated.Quantity, &updated.TotalAmount, &updated.Status)

tmpl := template.Must(template.ParseFiles("templates/status_updated.html"))
tmpl.Execute(w, updated)

	}
}

func deleteOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT order_id, customer_id, size, quantity, total_amount, status FROM orders")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var allOrders []Order
		for rows.Next() {
			var o Order
			rows.Scan(&o.OrderID, &o.CustomerID, &o.Size, &o.Quantity, &o.TotalAmount, &o.Status)
			allOrders = append(allOrders, o)
		}

		tmpl := template.Must(template.ParseFiles("templates/delete_order_form.html"))
		tmpl.Execute(w, allOrders)

	} else if r.Method == http.MethodPost {
		id := r.FormValue("orderid")

		result, err := db.Exec("DELETE FROM orders WHERE order_id = ?", id)
		if err != nil {
			http.Error(w, "Database delete error", http.StatusInternalServerError)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			tmpl := template.Must(template.ParseFiles("templates/order_not_found.html"))
			tmpl.Execute(w, nil)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/order_deleted.html"))
		tmpl.Execute(w, struct{ OrderID string }{OrderID: id})
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
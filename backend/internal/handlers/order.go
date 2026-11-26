package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderItem struct {
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}

func CreateOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Customer string      `json:"customer" binding:"required"`
			Items    []OrderItem `json:"items" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		skuMap := make(map[string]bool)
		for _, item := range req.Items {
			if skuMap[item.SKU] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate SKU in order: " + item.SKU})
				return
			}
			skuMap[item.SKU] = true
			if item.Qty <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than 0"})
				return
			}
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		totalPrice := 0
		var orderID int

		err = tx.QueryRow("INSERT INTO sales_orders (customer, status, total) VALUES ($1, 'CREATED', 0) RETURNING id", req.Customer).Scan(&orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, item := range req.Items {
			var productID int
			var currentStock, price int

			err = tx.QueryRow("SELECT id, stock, price FROM products WHERE sku = $1 FOR UPDATE", item.SKU).Scan(&productID, &currentStock, &price)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "SKU not found: " + item.SKU})
				return
			}

			if currentStock < item.Qty {
				tx.Rollback()
				c.JSON(http.StatusConflict, gin.H{
					"error": map[string]interface{}{
						"code":      "INSUFFICIENT_STOCK",
						"sku":       item.SKU,
						"available": currentStock,
					},
				})
				return
			}

			_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Qty, productID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			_, err = tx.Exec("INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES ($1, $2, $3, $4)", orderID, item.SKU, item.Qty, price)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			totalPrice += price * item.Qty
		}

		_, err = tx.Exec("UPDATE sales_orders SET total = $1 WHERE id = $2", totalPrice, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tx.Commit()
		c.JSON(http.StatusCreated, gin.H{"id": orderID, "total": totalPrice, "message": "Order created"})
	}
}

func CancelOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		var status string
		err = tx.QueryRow("SELECT status FROM sales_orders WHERE id = $1 FOR UPDATE", id).Scan(&status)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		if status == "CANCELLED" {
			c.JSON(http.StatusConflict, gin.H{"error": "Order already cancelled"})
			return
		}

		// Ambil semua item untuk dikembalikan stoknya
		rows, err := tx.Query("SELECT sku, qty FROM sales_order_items WHERE order_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Kita simpan dulu item-itemnya dalam slice
		var itemsToRestore []OrderItem
		for rows.Next() {
			var item OrderItem
			if err := rows.Scan(&item.SKU, &item.Qty); err != nil {
				rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			itemsToRestore = append(itemsToRestore, item)
		}
		rows.Close() // Tutup rows sebelum eksekusi update

		// Lakukan restore stok
		for _, item := range itemsToRestore {
			_, err = tx.Exec("UPDATE products SET stock = stock + $1 WHERE sku = $2", item.Qty, item.SKU)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		_, err = tx.Exec("UPDATE sales_orders SET status = 'CANCELLED', cancelled_at = now() WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Order cancelled and stock restored"})
	}
}

func GetOrder(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var order struct {
			ID        int    `json:"id"`
			Customer  string `json:"customer"`
			Status    string `json:"status"`
			Total     int    `json:"total"`
			CreatedAt string `json:"created_at"`
		}

		err := db.QueryRow("SELECT id, customer, status, total, created_at FROM sales_orders WHERE id = $1", id).Scan(
			&order.ID, &order.Customer, &order.Status, &order.Total, &order.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		rows, err := db.Query("SELECT sku, qty, unit_price FROM sales_order_items WHERE order_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var items []OrderItem
		for rows.Next() {
			var item OrderItem
			if err := rows.Scan(&item.SKU, &item.Qty, &item.UnitPrice); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         order.ID,
			"customer":   order.Customer,
			"status":     order.Status,
			"total":      order.Total,
			"created_at": order.CreatedAt,
			"items":      items,
		})
	}
}

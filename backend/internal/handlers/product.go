package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ID             int     `json:"id"`
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	Price          int     `json:"price"`
	Stock          int     `json:"stock"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	LastMovementAt *string `json:"last_movement_at"`
}

type Movement struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Type      string `json:"type"`
	Qty       int    `json:"qty"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

func GetProducts(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		offset := (page - 1) * limit

		query := `SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at, MAX(sm.created_at) as last_movement_at
				  FROM products p
				  LEFT JOIN stock_movements sm ON p.id = sm.product_id`

		args := []interface{}{}
		argCount := 1

		if search != "" {
			query += fmt.Sprintf(" WHERE p.sku ILIKE $%d OR p.name ILIKE $%d", argCount, argCount+1)
			args = append(args, "%"+search+"%", "%"+search+"%")
			argCount += 2
		}

		query += ` GROUP BY p.id ORDER BY p.id ASC LIMIT $` + strconv.Itoa(argCount) + ` OFFSET $` + strconv.Itoa(argCount+1)
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &p.LastMovementAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			products = append(products, p)
		}

		var total int
		countQuery := "SELECT COUNT(*) FROM products"
		if search != "" {
			countQuery += " WHERE sku ILIKE $1 OR name ILIKE $2"
			db.QueryRow(countQuery, "%"+search+"%", "%"+search+"%").Scan(&total)
		} else {
			db.QueryRow(countQuery).Scan(&total)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"page":  page,
			"limit": limit,
			"total": total,
		})
	}
}

func CreateProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SKU   string `json:"sku" binding:"required"`
			Name  string `json:"name" binding:"required"`
			Price int    `json:"price" binding:"required,min=0"`
			Stock int    `json:"stock"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id int
		err := db.QueryRow(
			"INSERT INTO products (sku, name, price, stock) VALUES ($1, $2, $3, $4) RETURNING id",
			req.SKU, req.Name, req.Price, req.Stock,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Product created"})
	}
}

func UpdateProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Name  string `json:"name"`
			Price int    `json:"price"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := db.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", req.Name, req.Price, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated"})
	}
}

func DeleteProduct(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var count int
		db.QueryRow("SELECT COUNT(*) FROM stock_movements WHERE product_id = $1", id).Scan(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": map[string]string{
				"code": "PRODUCT_HAS_HISTORY",
			}})
			return
		}

		res, err := db.Exec("DELETE FROM products WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

func AdjustStock(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Type string `json:"type" binding:"required,oneof=in out"`
			Qty  int    `json:"qty" binding:"required,gt=0"`
			Note string `json:"note"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		var currentStock int
		err = tx.QueryRow("SELECT stock FROM products WHERE id = $1 FOR UPDATE", id).Scan(&currentStock)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		newStock := currentStock
		if req.Type == "in" {
			newStock += req.Qty
		} else {
			newStock -= req.Qty
		}

		if newStock < 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": map[string]interface{}{
					"code":      "INSUFFICIENT_STOCK",
					"available": currentStock,
				},
			})
			return
		}

		_, err = tx.Exec("UPDATE products SET stock = $1 WHERE id = $2", newStock, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = tx.Exec("INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4)", id, req.Type, req.Qty, req.Note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Stock adjusted", "new_stock": newStock})
	}
}

func GetMovements(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		rows, err := db.Query("SELECT id, product_id, type, qty, note, created_at FROM stock_movements WHERE product_id = $1 ORDER BY created_at DESC", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var movements []Movement
		for rows.Next() {
			var m Movement
			err := rows.Scan(&m.ID, &m.ProductID, &m.Type, &m.Qty, &m.Note, &m.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			movements = append(movements, m)
		}

		c.JSON(http.StatusOK, gin.H{"data": movements})
	}
}

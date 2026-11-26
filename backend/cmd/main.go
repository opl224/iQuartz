package main

import (
	"log"
	"stockflow-backend/internal/config"
	"stockflow-backend/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.ConnectDB()
	defer db.Close()

	r := gin.Default()

	// --- TAMBAHKAN INI UNTUK CORS ---
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(config))
	// --------------------------------

	api := r.Group("/api")
	{
		api.GET("/products", handlers.GetProducts(db))
		api.POST("/products", handlers.CreateProduct(db))
		api.PUT("/products/:id", handlers.UpdateProduct(db))
		api.DELETE("/products/:id", handlers.DeleteProduct(db))
		api.POST("/products/:id/stock", handlers.AdjustStock(db))
		api.GET("/products/:id/movements", handlers.GetMovements(db))

		// Order routes (placeholder)
		api.POST("/orders", handlers.CreateOrder(db))
		api.POST("/orders/:id/cancel", handlers.CancelOrder(db))
		api.GET("/orders/:id", handlers.GetOrder(db))
	}

	log.Println("Server is running on port 8080...")
	r.Run(":8080")
}
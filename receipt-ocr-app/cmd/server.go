package main

import (
	"log"
	"receipt-ocr-app/database"
	"receipt-ocr-app/routes"

	// "receipt-ocr-app/backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database.InitDB()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Your React app
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Crud routes
	routes.CRUDRoutes(router)
	// routes.SetupRoutes(router)

	// Serve static files from uploads (optional)
	// router.Static("/uploads", "./uploads")

	// Start server
	log.Println("🚀 Server running on http://localhost:8000")
	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

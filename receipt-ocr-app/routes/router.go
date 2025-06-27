package routes

import (
	"net/http"
	"receipt-ocr-app/controllers"

	"github.com/gin-gonic/gin"
)

// CRUDRoutes sets up all the API endpoints
func CRUDRoutes(router *gin.Engine) {
	// Root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	router.POST("/upload", controllers.UploadReceipt)

	router.POST("/validate", controllers.ValidateReceipt)

	router.POST("/process", controllers.ProcessReceipt)

	router.GET("/receipts", controllers.GetAllReceipts)

	router.GET("/receipts/:id", controllers.GetReceiptById)

}

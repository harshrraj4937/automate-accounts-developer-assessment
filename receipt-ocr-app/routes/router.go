package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"receipt-ocr-app/database"
	"time"

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

	router.POST("/upload", func(c *gin.Context) {
		// Retrieve the file from form input (key: "file")
		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		// Validate that it is a PDF
		if filepath.Ext(file.Filename) != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
			return
		}

		// Ensure the uploads/ directory exisits
		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
			return
		}

		// Ensure unique filename with epoch
		// epoch := time.Now().Unix()
		// originalName := file.Filename
		// newFileName := fmt.Sprintf("%d_%s", epoch, originalName)
		// build file path to save
		filePath := filepath.Join("uploads", file.Filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// insert into receipt_file table
		query := `
			INSERT INTO receipt_file (
			file_name, file_path, created_at, updated_at
			) VALUES (?, ?, ?, ?)
		`
		now := time.Now().Format(time.RFC3339)
		_, err = database.DB.Exec(query, file.Filename, filePath, now, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata to database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "File uploaded successfully",
			"file_name": file.Filename,
			"file_path": filePath,
		})

	})

	router.POST("/validate", func(c *gin.Context) {
		var body struct {
			FileID   int    `json:"file_id"`
			FileName string `json:"file_name"`
		}

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// fetching the matching file record
		var filePath string
		query := "SELECT file_path FROM receipt_file WHERE id = ? AND file_name = ?"

		err := database.DB.QueryRow(query, body.FileID, body.FileName).Scan(&filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file with give ID and name does not exists"})
			return
		}

		// Ensurethis is the latest file upload with the same name
		var maxID int
		err = database.DB.QueryRow("SELECT MAX(id) FROM receipt_file WHERE file_name = ?", body.FileName).Scan(&maxID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch the latest file id"})
			return
		}
		if body.FileID != maxID {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Outdated file_id. Please validate the latest uploaded version of this file."})
			return
		}

		// check if the file exists
		file, err := os.Open(filePath)
		if err != nil {
			updateValidationStatus(body.FileID, false, "File not found on disk")
			c.JSON(http.StatusNotFound, gin.H{
				"file_id":        body.FileID,
				"file_name":      body.FileName,
				"is_valid":       false,
				"invalid_reason": "File not found on disk",
			})
			return
		}
		defer file.Close()

		// Check PDF header
		buf := make([]byte, 5)
		_, err = file.Read(buf)
		if err != nil || string(buf) != "%PDF-" {
			updateValidationStatus(body.FileID, false, "Invalid PDF format")
			c.JSON(http.StatusOK, gin.H{
				"file_id":        body.FileID,
				"file_name":      body.FileName,
				"is_valid":       false,
				"invalid_reason": "Invalid PDF format",
			})
			return
		}

		// file is valid
		updateValidationStatus(body.FileID, true, "")
		c.JSON(http.StatusOK, gin.H{
			"file_id":   body.FileID,
			"file_name": body.FileName,
			"is_valid":  true,
		})

	})

	// Other routes will be added here (e.g., /upload, /validate, etc.)

	router.POST("/process", func(c *gin.Context) {

		var body struct {
			FileID   int    `json:"file_id" binding:"required"`
			FileName string `json:"file_name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		// Fetch the file record to verify existence and get path
		var filePath string
		var isValid bool
		var isProcessed bool

		query := "SELECT file_path, is_valid, is_processed FROM receipt_file WHERE id = ? AND file_name = ?"
		err := database.DB.QueryRow(query, body.FileID, body.FileName).Scan(&filePath, &isValid, &isProcessed)

		if err != nil {
			fmt.Print(err)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found in database",
			})
			fmt.Print(err)
			return
		}

		// check file exists on disk
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File does not exists on disk"})
			return
		}
		// check if the file is valid
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File is not valid. Please validate it before processing.",
			})
			return
		}

		// if the file is already processed return the dummy response
		if isProcessed {
			var merchantName string
			var totalAmount float64
			var purchasedAt string

			selectQuery := `
				SELECT merchant_name, total_amount, purchased_at
				FROM receipt WHERE file_path = ?
				LIMIT 1
			`
			err := database.DB.QueryRow(selectQuery, filePath).Scan(&merchantName, &totalAmount, &purchasedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Receipt already processed, but failed to fetch data from the receipt table",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":       "Receipt was already processed",
				"merchant_name": merchantName,
				"total_amount":  totalAmount,
				"purchased_at":  purchasedAt,
			})
			return
		}

		// inserting dummy data for now (simulate the OCR)
		purchasedAt := "2025-06-27T15:04:05Z"
		merchantName := "Dummy Store"
		totalAmount := 100.00
		now := time.Now().Format(time.RFC3339)

		// insert into receipt table
		insert := `
			INSERT INTO receipt (purchased_at, merchant_name, total_amount, file_path, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`

		_, err = database.DB.Exec(insert, purchasedAt, merchantName, totalAmount, filePath, now, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into receipt table"})
			return
		}

		// update the receipt_file to mark as processed
		update := `
			UPDATE receipt_file SET is_processed = ?, updated_at = ?
			WHERE id = ?
		`

		_, err = database.DB.Exec(update, true, now, body.FileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update receipt_file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"messagge":      "Receipt processed successfully",
			"merchant_name": merchantName,
			"total_amount":  totalAmount,
			"purchased_at":  purchasedAt,
		})

	})
}

// this is a utility function so add it in the utility folder
func updateValidationStatus(fileID int, isValid bool, reason string) {
	query := `
		UPDATE receipt_file
		SET is_valid = ?, invalid_reason = ?, updated_at = ?
		WHERE id = ?
	`
	database.DB.Exec(query, isValid, reason, time.Now().Format(time.RFC3339), fileID)
}

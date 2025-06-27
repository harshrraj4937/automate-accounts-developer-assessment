package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"receipt-ocr-app/constants"
	"receipt-ocr-app/database"
	"time"

	"github.com/gin-gonic/gin"
)

type Receipt struct {
	ID           int     `json:"id"`
	PurchasedAt  string  `json:"purchased_at"`
	MerchantName string  `json:"merchant_name"`
	TotalAmount  float64 `json:"total_amount"`
	FilePath     string  `json:"file_path"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func UploadReceipt(c *gin.Context) {
	// UploadReceipt handles uploading a scanned receipt in PDF format.
	// Stores the file in the `uploads/` directory and saves metadata in the receipt_file table.
	//
	// Request packet:
	// curl --location --request POST 'http://localhost:8000/upload' \
	// --form 'file=@"/path/to/receipt.pdf"'
	//
	// Response packet:
	// {
	//     "message": "File uploaded successfully",
	//     "file_id": 3,
	//     "file_name": "receipt_1719502227.pdf",
	//     "file_path": "uploads/receipt_1719502227.pdf"
	// }

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

	now := time.Now().Format(time.RFC3339)

	result, err := database.DB.Exec(constants.InsertIntoReceiptFile, file.Filename, filePath, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata to database"})
		return
	}
	fileID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch inserted file ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"file_id":   fileID,
		"file_name": file.Filename,
		"file_path": filePath,
	})
}

func ValidateReceipt(c *gin.Context) {
	// ValidateReceipt checks whether a given file is a valid PDF.
	// Updates `is_valid` and `invalid_reason` fields in the receipt_file table.
	//
	// Request packet:
	// curl --location 'http://localhost:8000/validate' \
	// --header 'Content-Type: application/json' \
	// --data '{
	//     "file_id": 3,
	//     "file_name": "receipt_1719502227.pdf"
	// }'
	//
	// Response packet (valid):
	// {
	//     "file_id": 3,
	//     "file_name": "receipt_1719502227.pdf",
	//     "is_valid": true
	// }
	//
	// Response packet (invalid):
	// {
	//     "file_id": 3,
	//     "file_name": "receipt_1719502227.pdf",
	//     "is_valid": false,
	//     "invalid_reason": "Invalid PDF format"
	// }

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
	// query := "SELECT file_path FROM receipt_file WHERE id = ? AND file_name = ?"

	err := database.DB.QueryRow(constants.GetReceiptFileByIDAndFileNmae, body.FileID, body.FileName).Scan(&filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file with give ID and name does not exists"})
		return
	}

	// Ensurethis is the latest file upload with the same name
	var maxID int
	err = database.DB.QueryRow(constants.GetReceiptFileOfMaxID, body.FileName).Scan(&maxID)
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

}

func ProcessReceipt(c *gin.Context) {
	// ProcessReceipt simulates extracting receipt data using OCR.
	// Dummy data is inserted into the `receipt` table and `is_processed` is marked true.
	//
	// Only processes if the file is valid and not already processed.
	//
	// Request packet:
	// curl --location 'http://localhost:8000/process' \
	// --header 'Content-Type: application/json' \
	// --data '{
	//     "file_id": 3,
	//     "file_name": "receipt_1719502227.pdf"
	// }'
	//
	// Response packet:
	// {
	//     "message": "Receipt processed successfully",
	//     "merchant_name": "Dummy Store",
	//     "total_amount": 100,
	//     "purchased_at": "2025-06-27T15:04:05Z"
	// }

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

	// query := "SELECT file_path, is_valid, is_processed FROM receipt_file WHERE id = ? AND file_name = ?"
	err := database.DB.QueryRow(constants.GetReceiptDataByIDAndFileName, body.FileID, body.FileName).Scan(&filePath, &isValid, &isProcessed)

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

		err := database.DB.QueryRow(constants.GetReceiptDataByFilePath, filePath).Scan(&merchantName, &totalAmount, &purchasedAt)
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

	_, err = database.DB.Exec(constants.InsertIntoReceipt, purchasedAt, merchantName, totalAmount, filePath, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert into receipt table"})
		return
	}

	// update the receipt_file to mark as processed

	_, err = database.DB.Exec(constants.UpdateReceipFileIsProcessedUpdatedAtByID, true, now, body.FileID)
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

}

func GetAllReceipts(c *gin.Context) {
	// GetAllReceipts returns a list of all processed receipts from the receipt table.
	//
	// Request packet:
	// curl --location 'http://localhost:8000/receipts'
	//
	// Response packet:
	// {
	//     "receipts": [
	//         {
	//             "id": 1,
	//             "purchased_at": "2025-06-27T15:04:05Z",
	//             "merchant_name": "Dummy Store",
	//             "total_amount": 100,
	//             "file_path": "uploads/receipt_1719502227.pdf",
	//             "created_at": "2025-06-27T17:15:21Z",
	//             "updated_at": "2025-06-27T17:15:21Z"
	//         }
	//     ]
	// }

	// Get all receipts from the receipt table based on off set should be limit
	rows, err := database.DB.Query(constants.GetALLReceipts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch receipts",
		})
		return
	}
	// closing the db connection
	defer rows.Close()
	type Receipt struct {
		ID           int     `json:"id"`
		PurchasedAt  string  `json:"purchased_at"`
		MerchantName string  `json:"merchant_name"`
		TotalAmount  float64 `json:"total_amount"`
		FilePath     string  `json:"file_path"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
	}
	var receipts []Receipt

	for rows.Next() {
		var r Receipt
		err := rows.Scan(&r.ID, &r.PurchasedAt, &r.MerchantName, &r.TotalAmount, &r.FilePath, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse receipt row"})
			return
		}
		receipts = append(receipts, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"receipts": receipts,
	})
}

func GetReceiptById(c *gin.Context) {
	// Get receipt obj from the receipt table based on id
	// request packet
	// curl --location 'http://localhost:8000/receipts/{id}/'
	// response packet
	// {
	//     "id": 2,
	//     "purchased_at": "2025-06-27T15:04:05Z",
	//     "merchant_name": "Dummy Store",
	//     "total_amount": 100,
	//     "file_path": "uploads/venetian_434280912998.pdf",
	//     "created_at": "2025-06-27T05:19:50+05:30",
	//     "updated_at": "2025-06-27T05:19:50+05:30"
	// }

	// get the ID from URL param
	id := c.Param("id")

	var receipt struct {
		ID           int     `json:"id"`
		PurchasedAt  string  `json:"purchased_at"`
		MerchantName string  `json:"merchant_name"`
		TotalAmount  float64 `json:"total_amount"`
		FilePath     string  `json:"file_path"`
		CreatedAt    string  `json:"created_at"`
		UpdatedAt    string  `json:"updated_at"`
	}

	err := database.DB.QueryRow(constants.GetReceiptByID, id).Scan(
		&receipt.ID,
		&receipt.PurchasedAt,
		&receipt.MerchantName,
		&receipt.TotalAmount,
		&receipt.FilePath,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	c.JSON(http.StatusOK, receipt)
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

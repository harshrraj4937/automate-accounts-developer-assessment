package constants

const (
	GetReceiptByID string = `SELECT id, purchased_at, merchant_name, total_amount, file_path, created_at, updated_at
								FROM receipt
								WHERE id = $1
							`
	// need to work on off set based on search query todo
	GetALLReceipts string = `SELECT id, purchased_at, merchant_name, total_amount, file_path, created_at, updated_at FROM receipt`

	GetReceiptDataByFilePath string = `SELECT merchant_name, total_amount, purchased_at
										FROM receipt WHERE file_path = $1
										LIMIT 1
										`

	GetReceiptDataByIDAndFileName string = `SELECT file_path, is_valid, is_processed FROM receipt_file WHERE id = $1 AND file_name = $2`

	UpdateReceipFileIsProcessedUpdatedAtByID string = `UPDATE receipt_file SET is_processed = $1, updated_at = $2
														WHERE id = $3
														`

	InsertIntoReceipt string = `INSERT INTO receipt (purchased_at, merchant_name, total_amount, file_path, created_at, updated_at)
								VALUES ($1, $2, $3, $4, $5, $5)
								`

	GetReceiptFileByIDAndFileNmae string = `SELECT file_path FROM receipt_file WHERE id = $1 AND file_name = $2`

	GetReceiptFileOfMaxID = `SELECT MAX(id) FROM receipt_file WHERE file_name = $1`

	InsertIntoReceiptFile = `INSERT INTO receipt_file (
							file_name, file_path, created_at, updated_at
							) VALUES ($1, $2, $3, $4)
							`
)

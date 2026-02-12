package archiver

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"time"

	"pos-fiber-app/internal/business"
	"pos-fiber-app/internal/sale"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// UploadBackupToDrive fetches transactions and uploads them as a CSV to Google Drive
func UploadBackupToDrive(db *gorm.DB, biz *business.Business, from, to time.Time) error {
	if !biz.GoogleDriveLinked || biz.GoogleDriveFolderID == "" {
		return fmt.Errorf("google drive not linked for business %s", biz.Name)
	}

	// 1. Refresh/Ensure token is valid
	token, err := business.RefreshGoogleToken(db, biz)
	if err != nil {
		return fmt.Errorf("failed to refresh google token: %w", err)
	}

	// 2. Fetch data to backup
	var sales []sale.Sale
	if err := db.Preload("SaleItems").Where("business_id = ? AND sale_date >= ? AND sale_date < ?", biz.ID, from, to).Find(&sales).Error; err != nil {
		return err
	}

	if len(sales) == 0 {
		log.Printf("[ARCHIVER] No sales found to backup for %s in period %s to %s\n", biz.Name, from.Format("2006-01-02"), to.Format("2006-01-02"))
		return nil
	}

	// 3. Generate CSV in memory
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// CSV Header
	writer.Write([]string{"SaleID", "Date", "Customer", "Total", "Payment", "Items"})

	for _, s := range sales {
		itemsDesc := ""
		for _, item := range s.SaleItems {
			itemsDesc += fmt.Sprintf("%dx %s; ", item.Quantity, item.ProductName)
		}
		writer.Write([]string{
			fmt.Sprintf("%d", s.ID),
			s.SaleDate.Format("2006-01-02 15:04:05"),
			s.CustomerName,
			fmt.Sprintf("%.2f", s.Total),
			s.PaymentMethod,
			itemsDesc,
		})
	}
	writer.Flush()

	// 4. Upload to Google Drive using SDK
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithTokenSource(business.GetTokenSource(token)))
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("Backup_%s_%s_to_%s.csv", biz.Name, from.Format("2006-01-02"), to.Format("2006-01-02"))

	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{biz.GoogleDriveFolderID},
	}

	_, err = srv.Files.Create(driveFile).Media(bytes.NewReader(buf.Bytes())).Do()
	if err != nil {
		return fmt.Errorf("failed to upload to drive: %w", err)
	}

	log.Printf("[ARCHIVER] Uploaded backup for %s to Google Drive: %s\n", biz.Name, fileName)
	return nil
}

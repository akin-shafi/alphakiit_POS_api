// internal/inventory/reservation.go
package inventory

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// StockReservation tracks reserved stock for draft/held orders
type StockReservation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductID  uint      `gorm:"index:idx_product_business" json:"product_id"`
	BusinessID uint      `gorm:"index:idx_product_business" json:"business_id"`
	SaleID     uint      `gorm:"index" json:"sale_id"` // Links to draft/held sale
	Quantity   int       `json:"quantity"`
	CashierID  uint      `json:"cashier_id"`
	ExpireAt   time.Time `json:"expire_at"` // Auto-release after X hours
	CreatedAt  time.Time `json:"created_at"`
}

// ReservationService handles stock reservation operations
type ReservationService struct {
	db *gorm.DB
}

// NewReservationService creates a new reservation service
func NewReservationService(db *gorm.DB) *ReservationService {
	return &ReservationService{db: db}
}

// ReserveStock reserves stock for a draft/held sale
func (s *ReservationService) ReserveStock(saleID, productID, businessID, cashierID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	// Check if sufficient stock is available (current - reserved)
	available, err := s.GetAvailableStock(productID, businessID)
	if err != nil {
		return err
	}

	if available < quantity {
		return errors.New("insufficient stock available for reservation")
	}

	// Create reservation with 4-hour expiry
	reservation := &StockReservation{
		ProductID:  productID,
		BusinessID: businessID,
		SaleID:     saleID,
		Quantity:   quantity,
		CashierID:  cashierID,
		ExpireAt:   time.Now().Add(4 * time.Hour),
	}

	return s.db.Create(reservation).Error
}

// ReleaseReservation releases stock reservation for a sale
func (s *ReservationService) ReleaseReservation(saleID, productID uint) error {
	result := s.db.Where("sale_id = ? AND product_id = ?", saleID, productID).Delete(&StockReservation{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ReleaseAllReservations releases all reservations for a specific sale
func (s *ReservationService) ReleaseAllReservations(saleID uint) error {
	return s.db.Where("sale_id = ?", saleID).Delete(&StockReservation{}).Error
}

// GetAvailableStock returns available stock (current - reserved)
func (s *ReservationService) GetAvailableStock(productID, businessID uint) (int, error) {
	// Get current stock
	var inv Inventory
	if err := s.db.First(&inv, "product_id = ? AND business_id = ?", productID, businessID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("product not found in inventory")
		}
		return 0, err
	}

	// Get total reserved stock
	var totalReserved int
	err := s.db.Model(&StockReservation{}).
		Where("product_id = ? AND business_id = ? AND expire_at > ?", productID, businessID, time.Now()).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalReserved).Error

	if err != nil {
		return 0, err
	}

	available := inv.CurrentStock - totalReserved
	if available < 0 {
		available = 0
	}

	return available, nil
}

// GetReservedStock returns total reserved stock for a product
func (s *ReservationService) GetReservedStock(productID, businessID uint) (int, error) {
	var totalReserved int
	err := s.db.Model(&StockReservation{}).
		Where("product_id = ? AND business_id = ? AND expire_at > ?", productID, businessID, time.Now()).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalReserved).Error

	return totalReserved, err
}

// CleanExpiredReservations removes reservations that have expired
// This should be called by a cron job
func (s *ReservationService) CleanExpiredReservations() error {
	result := s.db.Where("expire_at < ?", time.Now()).Delete(&StockReservation{})
	if result.Error != nil {
		return result.Error
	}

	// Log how many were cleaned
	if result.RowsAffected > 0 {
		// TODO: Add logging here
		// log.Printf("Cleaned %d expired reservations", result.RowsAffected)
	}

	return nil
}

// GetReservationsBySale returns all reservations for a specific sale
func (s *ReservationService) GetReservationsBySale(saleID uint) ([]StockReservation, error) {
	var reservations []StockReservation
	err := s.db.Where("sale_id = ?", saleID).Find(&reservations).Error
	return reservations, err
}

// UpdateReservationQuantity updates the quantity of an existing reservation
func (s *ReservationService) UpdateReservationQuantity(saleID, productID uint, newQuantity int) error {
	if newQuantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	if newQuantity == 0 {
		// If quantity is 0, delete the reservation
		return s.ReleaseReservation(saleID, productID)
	}

	var reservation StockReservation
	if err := s.db.First(&reservation, "sale_id = ? AND product_id = ?", saleID, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("reservation not found")
		}
		return err
	}

	// Check if sufficient stock is available for the new quantity
	// (current available + old reservation quantity)
	available, err := s.GetAvailableStock(productID, reservation.BusinessID)
	if err != nil {
		return err
	}

	availableWithOld := available + reservation.Quantity
	if availableWithOld < newQuantity {
		return errors.New("insufficient stock available for new quantity")
	}

	reservation.Quantity = newQuantity
	reservation.ExpireAt = time.Now().Add(4 * time.Hour) // Reset expiry

	return s.db.Save(&reservation).Error
}

// MigrateReservations runs the database migration for reservations
func MigrateReservations(db *gorm.DB) error {
	return db.AutoMigrate(&StockReservation{})
}

// internal/product/service.go
package product

import (
	"errors"

	"gorm.io/gorm"
)

// Create creates a new product for the given business
func Create(db *gorm.DB, businessID uint, req CreateProductRequest) (*Product, error) {
	product := &Product{
		BusinessID:  businessID,
		Name:        req.Name,
		SKU:         req.SKU,
		Description: req.Description,
		Price:       req.Price,
		Cost:        req.Cost,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
		Stock:       req.Stock,
		MinStock:    req.MinStock,
		Barcode:     req.Barcode,
		Active:      true, // default
	}

	if err := db.Create(product).Error; err != nil {
		return nil, err
	}

	// Create inventory record
	inv := struct {
		ProductID    uint `gorm:"column:product_id"`
		BusinessID   uint `gorm:"column:business_id"`
		CurrentStock int  `gorm:"column:current_stock"`
	}{
		ProductID:    product.ID,
		BusinessID:   businessID,
		CurrentStock: product.Stock,
	}
	db.Table("inventories").Create(&inv)

	return product, nil
}

// ListByBusiness returns all active products for a business, optionally filtered
func ListByBusiness(db *gorm.DB, businessID uint, filters ...func(*gorm.DB) *gorm.DB) ([]Product, error) {
	var products []Product

	query := db.Table("products").
		Select("products.id, products.business_id, products.category_id, products.name, products.sku, products.description, products.price, products.cost, products.image_url, products.min_stock, products.barcode, products.active, COALESCE(inventories.current_stock, products.stock) as stock").
		Joins("LEFT JOIN inventories ON inventories.product_id = products.id AND inventories.business_id = products.business_id").
		Where("products.business_id = ? AND products.active = ?", businessID, true)

	for _, filter := range filters {
		query = filter(query)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// Optional filter helpers (you can expand)
func WithCategory(categoryID uint) func(*gorm.DB) *gorm.DB {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where("category_id = ?", categoryID)
	}
}

func WithSearch(term string) func(*gorm.DB) *gorm.DB {
	return func(q *gorm.DB) *gorm.DB {
		term = "%" + term + "%"
		return q.Where("name ILIKE ? OR sku ILIKE ?", term, term)
	}
}

// Get retrieves a single product by ID, ensuring it belongs to the business
func Get(db *gorm.DB, id, businessID uint) (*Product, error) {
	var product Product

	err := db.Table("products").
		Select("products.id, products.business_id, products.category_id, products.name, products.sku, products.description, products.price, products.cost, products.image_url, products.min_stock, products.barcode, products.active, COALESCE(inventories.current_stock, products.stock) as stock").
		Joins("LEFT JOIN inventories ON inventories.product_id = products.id AND inventories.business_id = products.business_id").
		Where("products.id = ? AND products.business_id = ?", id, businessID).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &product, nil
}

// Update modifies an existing product (partial updates allowed)
func Update(db *gorm.DB, id, businessID uint, req UpdateProductRequest) (*Product, error) {
	product, err := Get(db, id, businessID)
	if err != nil {
		return nil, err
	}

	// Apply updates only if fields are provided
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.SKU != "" {
		product.SKU = req.SKU
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Cost != nil {
		product.Cost = *req.Cost
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.MinStock != nil {
		product.MinStock = *req.MinStock
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.Active != nil {
		product.Active = *req.Active
	}

	if err := db.Save(product).Error; err != nil {
		return nil, err
	}

	// Sync to inventory table
	if req.Stock != nil {
		var count int64
		db.Table("inventories").Where("product_id = ? AND business_id = ?", id, businessID).Count(&count)
		if count > 0 {
			db.Table("inventories").Where("product_id = ? AND business_id = ?", id, businessID).Update("current_stock", *req.Stock)
		} else {
			inv := struct {
				ProductID    uint `gorm:"column:product_id"`
				BusinessID   uint `gorm:"column:business_id"`
				CurrentStock int  `gorm:"column:current_stock"`
			}{
				ProductID:    id,
				BusinessID:   businessID,
				CurrentStock: *req.Stock,
			}
			db.Table("inventories").Create(&inv)
		}
	}

	return product, nil
}

// Delete soft-deletes a product by setting Active = false
// (Better than hard delete for audit/sales history)
func Delete(db *gorm.DB, id, businessID uint) error {
	product, err := Get(db, id, businessID)
	if err != nil {
		return err
	}

	product.Active = false

	return db.Save(product).Error
}

// Optional: HardDelete if needed (use cautiously)
// func HardDelete(db *gorm.DB, id, businessID uint) error {
//     return db.Where("id = ? AND business_id = ?", id, businessID).Delete(&Product{}).Error
// }

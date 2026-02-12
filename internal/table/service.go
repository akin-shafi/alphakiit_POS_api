// internal/table/service.go
package table

import (
	"errors"

	"gorm.io/gorm"
)

type TableService struct {
	db *gorm.DB
}

func NewTableService(db *gorm.DB) *TableService {
	return &TableService{db: db}
}

// CreateTable creates a new table for a business
func (s *TableService) CreateTable(businessID uint, tableNumber, section string, capacity int) (*Table, error) {
	// Check if table number already exists for this business
	var existing Table
	err := s.db.Where("business_id = ? AND table_number = ?", businessID, tableNumber).First(&existing).Error
	if err == nil {
		return nil, errors.New("table number already exists")
	}

	if capacity <= 0 {
		capacity = 4 // Default capacity
	}

	table := &Table{
		BusinessID:  businessID,
		TableNumber: tableNumber,
		Section:     section,
		Capacity:    capacity,
		Status:      StatusAvailable,
	}

	if err := s.db.Create(table).Error; err != nil {
		return nil, err
	}

	return table, nil
}

// ListTables returns all tables for a business
func (s *TableService) ListTables(businessID uint, section string) ([]Table, error) {
	var tables []Table
	query := s.db.Where("business_id = ?", businessID)

	if section != "" {
		query = query.Where("section = ?", section)
	}

	err := query.Order("table_number ASC").Find(&tables).Error
	return tables, err
}

// GetTable retrieves a specific table
func (s *TableService) GetTable(tableID, businessID uint) (*Table, error) {
	var table Table
	err := s.db.First(&table, "id = ? AND business_id = ?", tableID, businessID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("table not found")
		}
		return nil, err
	}
	return &table, nil
}

// GetTableByNumber retrieves a table by its number
func (s *TableService) GetTableByNumber(businessID uint, tableNumber string) (*Table, error) {
	var table Table
	err := s.db.First(&table, "business_id = ? AND table_number = ?", businessID, tableNumber).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("table not found")
		}
		return nil, err
	}
	return &table, nil
}

// UpdateTable updates table information
func (s *TableService) UpdateTable(tableID, businessID uint, tableNumber, section string, capacity int, status TableStatus) (*Table, error) {
	table, err := s.GetTable(tableID, businessID)
	if err != nil {
		return nil, err
	}

	// If changing table number, check for duplicates
	if tableNumber != "" && tableNumber != table.TableNumber {
		var existing Table
		err := s.db.Where("business_id = ? AND table_number = ? AND id != ?", businessID, tableNumber, tableID).First(&existing).Error
		if err == nil {
			return nil, errors.New("table number already exists")
		}
		table.TableNumber = tableNumber
	}

	if section != "" {
		table.Section = section
	}

	if capacity > 0 {
		table.Capacity = capacity
	}

	if status != "" {
		table.Status = status
	}

	if err := s.db.Save(table).Error; err != nil {
		return nil, err
	}

	return table, nil
}

// UpdateTableStatus updates only the status of a table
func (s *TableService) UpdateTableStatus(tableID, businessID uint, status TableStatus) error {
	result := s.db.Model(&Table{}).
		Where("id = ? AND business_id = ?", tableID, businessID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// DeleteTable deletes a table
func (s *TableService) DeleteTable(tableID, businessID uint) error {
	// TODO: Check if table has active orders before deleting
	// For now, we'll allow deletion

	result := s.db.Where("id = ? AND business_id = ?", tableID, businessID).Delete(&Table{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// GetTableOrders returns information about a table including active orders
// This will be implemented when we integrate with the sales service
func (s *TableService) GetTableOrders(tableID, businessID uint) (*TableWithOrders, error) {
	table, err := s.GetTable(tableID, businessID)
	if err != nil {
		return nil, err
	}

	// TODO: Query sales table for draft/held orders for this table
	// For now, return table with zero orders
	return &TableWithOrders{
		Table:        *table,
		ActiveOrders: 0,
		TotalAmount:  0,
	}, nil
}

// GetSectionList returns unique sections for a business
func (s *TableService) GetSectionList(businessID uint) ([]string, error) {
	var sections []string
	err := s.db.Model(&Table{}).
		Where("business_id = ? AND section != ''", businessID).
		Distinct("section").
		Pluck("section", &sections).Error

	return sections, err
}

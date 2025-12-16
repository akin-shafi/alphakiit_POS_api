// internal/category/service.go
package category

import "gorm.io/gorm"

func Create(db *gorm.DB, businessID uint, name, desc string) (*Category, error) {
	cat := &Category{BusinessID: businessID, Name: name, Description: desc}
	return cat, db.Create(cat).Error
}

func ListByBusiness(db *gorm.DB, businessID uint) ([]Category, error) {
	var cats []Category
	err := db.Where("business_id = ?", businessID).Find(&cats).Error
	return cats, err
}

func Get(db *gorm.DB, id, businessID uint) (*Category, error) {
	var cat Category
	err := db.Where("id = ? AND business_id = ?", id, businessID).First(&cat).Error
	return &cat, err
}

func Update(db *gorm.DB, id, businessID uint, name, desc string) (*Category, error) {
	cat, err := Get(db, id, businessID)
	if err != nil {
		return nil, err
	}
	cat.Name = name
	cat.Description = desc
	return cat, db.Save(cat).Error
}

func Delete(db *gorm.DB, id, businessID uint) error {
	return db.Where("id = ? AND business_id = ?", id, businessID).Delete(&Category{}).Error
}
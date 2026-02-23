package advert

import (
	"errors"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateAdvert(req CreateAdvertRequest) (*Advert, error) {
	advert := Advert{
		BusinessID: req.BusinessID,
		Title:      req.Title,
		Type:       req.Type,
		URL:        req.URL,
		Active:     req.Active,
	}

	if err := s.db.Create(&advert).Error; err != nil {
		return nil, err
	}
	return &advert, nil
}

func (s *Service) GetAdverts(businessID *uint) ([]Advert, error) {
	var adverts []Advert
	query := s.db.Where("active = ?", true)
	if businessID != nil {
		query = query.Where("business_id = ? OR business_id IS NULL", *businessID)
	} else {
		query = query.Where("business_id IS NULL")
	}

	if err := query.Find(&adverts).Error; err != nil {
		return nil, err
	}
	return adverts, nil
}

func (s *Service) GetAdvertByID(id uint) (*Advert, error) {
	var advert Advert
	if err := s.db.First(&advert, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("advert not found")
		}
		return nil, err
	}
	return &advert, nil
}

func (s *Service) UpdateAdvert(id uint, req UpdateAdvertRequest) (*Advert, error) {
	advert, err := s.GetAdvertByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		advert.Title = *req.Title
	}
	if req.Type != nil {
		advert.Type = *req.Type
	}
	if req.URL != nil {
		advert.URL = *req.URL
	}
	if req.Active != nil {
		advert.Active = *req.Active
	}

	if err := s.db.Save(advert).Error; err != nil {
		return nil, err
	}
	return advert, nil
}

func (s *Service) DeleteAdvert(id uint) error {
	if err := s.db.Delete(&Advert{}, id).Error; err != nil {
		return err
	}
	return nil
}

package user

import (
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Create a new user under a tenant
func (s *UserService) Create(user *User) error {
	// Check if email already exists
	var count int64
	s.db.Model(&User{}).Where("email = ?", user.Email).Count(&count)
	if count > 0 {
		return errors.New("email already in use")
	}

	if user.Password != "" {
		hashed, err := HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashed
	}
	return s.db.Create(user).Error
}

// ListByTenant returns all users for a given tenant
func (s *UserService) ListByTenant(tenantID string) ([]User, error) {
	var users []User
	err := s.db.Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

// GetByID retrieves a single user by ID
func (s *UserService) GetByID(id uint) (*User, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user
func (s *UserService) Update(id uint, data *User) (*User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if data.FirstName != "" {
		user.FirstName = data.FirstName
	}
	if data.LastName != "" {
		user.LastName = data.LastName
	}
	if data.Email != "" && data.Email != user.Email {
		var count int64
		s.db.Model(&User{}).Where("email = ? AND id != ?", data.Email, id).Count(&count)
		if count > 0 {
			return nil, errors.New("email already in use")
		}
		user.Email = data.Email
	}
	if data.Role != "" {
		user.Role = data.Role
	}
	if data.OutletID != nil {
		user.OutletID = data.OutletID
	}
	if data.Password != "" {
		hashed, err := HashPassword(data.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}

	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Delete removes a user
func (s *UserService) Delete(id uint) error {
	user, err := s.GetByID(id)
	if err != nil {
		return err
	}
	return s.db.Delete(user).Error
}

// ResetPassword updates a user's password
func (s *UserService) ResetPassword(id uint, password string) error {
	user, err := s.GetByID(id)
	if err != nil {
		return err
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hashed

	return s.db.Save(user).Error
}

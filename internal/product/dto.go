package product

// type CreateProductRequest struct {
// 	Name        string  `json:"name" validate:"required,min=2"`
// 	SKU         string  `json:"sku" validate:"omitempty,alphanum"`
// 	Description string  `json:"description"`
// 	Price       float64 `json:"price" validate:"required,gt=0"`
// 	Cost        float64 `json:"cost" validate:"gte=0"`
// 	CategoryID  uint    `json:"category_id" validate:"required"`
// 	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
// 	Active      bool    `json:"active" default:"true"`
// }

// type UpdateProductRequest struct {
// 	Name        string  `json:"name,omitempty"`
// 	SKU         string  `json:"sku,omitempty"`
// 	Description string  `json:"description,omitempty"`
// 	Price       float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
// 	Cost        float64 `json:"cost,omitempty" validate:"omitempty,gte=0"`
// 	CategoryID  uint    `json:"category_id,omitempty"`
// 	ImageURL    string  `json:"image_url,omitempty" validate:"omitempty,url"`
// 	Active      *bool   `json:"active,omitempty"`
// }

// CreateProductRequest
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	SKU         string  `json:"sku" validate:"omitempty,alphanum"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Cost        float64 `json:"cost" validate:"gte=0"`
	CategoryID  uint    `json:"category_id" validate:"required"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
}

// UpdateProductRequest (all fields optional)
type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty"`
	SKU         string   `json:"sku,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Cost        *float64 `json:"cost,omitempty" validate:"omitempty,gte=0"`
	CategoryID  *uint    `json:"category_id,omitempty"`
	ImageURL    string   `json:"image_url,omitempty" validate:"omitempty,url"`
	Active      *bool    `json:"active,omitempty"`
}

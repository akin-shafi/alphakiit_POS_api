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
	Name        string  `json:"name" form:"name" validate:"required,min=2"`
	SKU         string  `json:"sku" form:"sku" validate:"omitempty,alphanum"`
	Description string  `json:"description,omitempty" form:"description"`
	Price       float64 `json:"price" form:"price" validate:"required,gt=0"`
	Cost        float64 `json:"cost" form:"cost" validate:"gte=0"`
	CategoryID  uint    `json:"category_id" form:"category_id" validate:"required"`
	ImageURL    string  `json:"image_url" form:"image_url" validate:"omitempty,url"`
	Stock       int     `json:"stock" form:"stock"`
	MinStock    int     `json:"min_stock" form:"min_stock"`
	Barcode     string  `json:"barcode,omitempty" form:"barcode"`
	TrackByRound bool   `json:"track_by_round" form:"track_by_round"`
	UnitOfMeasure string `json:"unit_of_measure,omitempty" form:"unit_of_measure"`
}

// UpdateProductRequest (all fields optional)
type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty" form:"name"`
	SKU         string   `json:"sku,omitempty" form:"sku"`
	Description string   `json:"description,omitempty" form:"description"`
	Price       *float64 `json:"price,omitempty" form:"price" validate:"omitempty,gt=0"`
	Cost        *float64 `json:"cost,omitempty" form:"cost" validate:"omitempty,gte=0"`
	CategoryID  *uint    `json:"category_id,omitempty" form:"category_id"`
	ImageURL    string   `json:"image_url,omitempty" form:"image_url" validate:"omitempty,url"`
	Stock       *int     `json:"stock,omitempty" form:"stock"`
	MinStock    *int     `json:"min_stock,omitempty" form:"min_stock"`
	Barcode     string   `json:"barcode,omitempty" form:"barcode"`
	Active      *bool    `json:"active,omitempty" form:"active"`
	TrackByRound *bool   `json:"track_by_round,omitempty" form:"track_by_round"`
	UnitOfMeasure string `json:"unit_of_measure,omitempty" form:"unit_of_measure"`
}

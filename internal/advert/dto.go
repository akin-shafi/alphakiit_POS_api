package advert

type CreateAdvertRequest struct {
	BusinessID *uint      `json:"business_id"`
	Title      string     `json:"title" validate:"max=200"`
	Type       AdvertType `json:"type" validate:"required,oneof=IMAGE VIDEO"`
	URL        string     `json:"url" validate:"required,url"`
	Active     bool       `json:"active"`
}

type UpdateAdvertRequest struct {
	Title  *string     `json:"title" validate:"omitempty,max=200"`
	Type   *AdvertType `json:"type" validate:"omitempty,oneof=IMAGE VIDEO"`
	URL    *string     `json:"url" validate:"omitempty,url"`
	Active *bool       `json:"active"`
}

package outlet

type CreateOutletRequest struct {
	Name    string `json:"name" example:"Main Branch"`
	Address string `json:"address" example:"12 Market Street"`
}

package dto

type ImportResult struct {
	Kind    string   `json:"kind"`
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

type HotelRoomResponse struct {
	RoomNumber string `json:"room_number"`
	RoomType   string `json:"room_type,omitempty"`
	Floor      string `json:"floor,omitempty"`
}

type HotelMenuCategoryResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type HotelMenuItemResponse struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Tag         string `json:"tag,omitempty"`
}

type HotelMenuResponse struct {
	Categories []HotelMenuCategoryResponse `json:"categories"`
	Items      []HotelMenuItemResponse     `json:"items"`
}

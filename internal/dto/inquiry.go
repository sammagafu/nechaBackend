package dto

type SubmitInquiryRequest struct {
	Type     string `json:"type" validate:"required"`
	Name     string `json:"name"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone"`
	Company  string `json:"company"`
	Role     string `json:"role"`
	Location string `json:"location"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Rooms    int    `json:"rooms"`
	Units    int    `json:"units"`
	Price    int    `json:"price"`
}

type InquiryResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Role      string `json:"role"`
	Location  string `json:"location"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
}

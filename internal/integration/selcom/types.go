package selcom

type APIResponse struct {
	Reference  string                   `json:"reference"`
	ResultCode string                   `json:"resultcode"`
	Result     string                   `json:"result"`
	Message    string                   `json:"message"`
	Data       []map[string]interface{} `json:"data"`
}

type CreateOrderMinimalInput struct {
	OrderID         string
	BuyerEmail      string
	BuyerName       string
	BuyerPhone      string
	Amount          int64
	Currency        string
	RedirectURL     string
	CancelURL       string
	WebhookURL      string
	BuyerRemarks    string
	MerchantRemarks string
	NoOfItems       int
}

type CheckoutResult struct {
	Reference         string
	PaymentGatewayURL string
	PaymentToken      string
	QR                string
}

type WebhookPayload struct {
	Result        string `json:"result"`
	ResultCode    string `json:"resultcode"`
	OrderID       string `json:"order_id"`
	TransID       string `json:"transid"`
	Reference     string `json:"reference"`
	Channel       string `json:"channel"`
	Amount        string `json:"amount"`
	Phone         string `json:"phone"`
	PaymentStatus string `json:"payment_status"`
}

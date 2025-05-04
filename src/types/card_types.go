package types

type CreateCardReq struct {
	PGPKey string `json:"pgp_key"`
}

type CreateCardRes struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	CreatedAt  string `json:"created_at"`
	CardNumber string `json:"card_number"`
	Expire     string `json:"expire"`
	CVV        string `json:"cvv"`
}

type CardRes struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type CardDetailsRes struct {
	ID         int64  `json:"id"`
	CardNumber string `json:"card_number"`
	Expire     string `json:"expire"`
}

type CardListRes struct {
	Cards []CardRes `json:"cards"`
}

type PaymentReq struct {
	CardID int64  `json:"card_id"`
	Amount string `json:"amount"`
	CVV    string `json:"cvv"`
	PGPKey string `json:"pgp_key"`
}

type PaymentRes struct {
	Success     bool   `json:"success"`
	PaymentID   string `json:"payment_id,omitempty"`
	Description string `json:"description,omitempty"`
}
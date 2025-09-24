package response

type QrLoginStatus string

const (
	StatusPending  QrLoginStatus = "PENDING"
	StatusApproved QrLoginStatus = "APPROVED"
	StatusExpired  QrLoginStatus = "EXPIRED"
)

type QrLoginResponse struct {
	Status QrLoginStatus `json:"status"`
	Tokens *Tokens       `json:"tokens,omitempty"`
}

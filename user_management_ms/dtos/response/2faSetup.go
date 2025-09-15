package response

type TwoFASetupResponse struct {
	Secret string `json:"secret"`
	QRCode []byte `json:"qr_code"` // optional
}

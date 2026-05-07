package model

type TextData struct {
	Text string `json:"text"`
}

type CardData struct {
	Number     string `json:"number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	HolderName string `json:"holder_name"`
}

type CredentialsData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
	Note     string `json:"note,omitempty"`
}

type TokenData struct {
	Token string `json:"token"`
	Login string `json:"login"`
}

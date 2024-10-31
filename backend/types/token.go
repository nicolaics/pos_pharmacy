package types

type TokenDetails struct {
	Token    string `json:"token"`
	UUID     string `json:"uuid"`
	TokenExp int64  `json:"tokenExp"`
}

type AccessDetails struct {
	UUID   string
	UserID int
}
package model

type InternalToken struct {
	Verified bool `json:"verified"`
	Jwt string `json:"jwt"`
}

type VerificationResp struct {
	Verified bool 
	Username string 
}
package protocol

type ReqUserLogin struct {
	Package    string `json:"package"`
	Version    string `json:"version"`
	Duid       string `json:"duid"`
	Credential string `json:"credential"`
	FCMToken   string `json:"token"`
	PID        int    `json:"pid"`
	P          int    `json:"p"`
}

type RespUserLogin struct {
	Credential string `json:"credential"`
	Coins      int64  `json:"coins"`
	UserID     int    `json:"user_id"`
	CommonError
}

package model

type Config struct {
	Redis        *Redis         `json:"redis"`
	Mongodb      *Mongodb       `json:"mongodb"`
	BlackListMap map[string]int `json:"-"`
	WhiteListMap map[string]int `json:"-"`
}

type Redis struct {
	Path         string   `json:"path"`
	SentinelPath []string `json:"sentinel"`
	Password     string   `json:"password"`
	DB           int      `json:"db"`
}

type Mongodb struct {
	Path        string `json:"path"`
	CaFilePath  string `json:"ca_file_path"`
	CertPemPath string `json:"cert_pem_path"`
	KeyPemPath  string `json:"key_pem_path"`
}

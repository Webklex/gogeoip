package server

type Tls struct {
	ServerAddr string `json:"server_addr"`
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
}

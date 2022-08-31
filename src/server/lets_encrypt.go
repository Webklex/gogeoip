package server

type LetsEncrypt struct {
	Enabled  bool   `json:"enabled"`
	CacheDir string `json:"cache_dir"`
	Email    string `json:"email"`
	Hosts    string `json:"hosts"`
}

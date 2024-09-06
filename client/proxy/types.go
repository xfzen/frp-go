package proxy

type ConfigInfo struct {
	Addr    string `json:"addr"`
	Inspect bool   `json:"inspect"`
}

type WorkingDetial struct {
	Name      string `json:"name"`
	Uri       string `json:"uri"`
	PublicUrl string `json:"public_url"`
	Type      string `json:"type"`
	Status    string `json:"status"`

	Config ConfigInfo `json:"config"`
}

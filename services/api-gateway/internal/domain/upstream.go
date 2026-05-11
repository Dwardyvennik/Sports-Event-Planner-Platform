package domain

type Upstream struct {
	Name    string
	Address string
}

type GatewayStatus struct {
	Service   string
	Upstreams []Upstream
}

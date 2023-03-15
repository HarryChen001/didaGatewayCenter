package deviceInfo

const (
	Method = "thing.deviceinfo.update"
)

type DeviceInfo struct {
	ID      string   `json:"id"`
	Version string   `json:"version"`
	Sys     Sys      `json:"sys,omitempty"`
	Params  []Params `json:"params"`
	Method  string   `json:"method"`
}

type Sys struct {
	Ack int `json:"ack"`
}

type Params struct {
	AttrKey   string `json:"attrKey"`
	AttrValue string `json:"attrValue"`
}

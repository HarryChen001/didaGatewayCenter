package ota

type UploadInfo struct {
	Id     string `json:"id"`
	Params Params `json:"params"`
}

type Params struct {
	Version string `json:"version"`
	Module  string `json:"module"`
}

type Packet struct {
	Code    string `json:"code"`
	Data    Data   `json:"data"`
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type ExtData struct {
	Key1       string `json:"key1"`
	Key2       string `json:"key2"`
	PackageUdi string `json:"_package_udi"`
}

type Data struct {
	Size       int     `json:"size"`
	Version    string  `json:"version"`
	IsDiff     int     `json:"isDiff"`
	URL        string  `json:"url"`
	Md5        string  `json:"md5"`
	DigestSign string  `json:"digestsign"`
	Sign       string  `json:"sign"`
	SignMethod string  `json:"signMethod"`
	Module     string  `json:"module"`
	ExtData    ExtData `json:"extData"`
}

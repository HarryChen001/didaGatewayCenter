package domain

type AppConfig struct {
	Debug  bool `json:"debug" yaml:"debug"`
	Server struct {
		Address string `json:"address" yaml:"address"`
	} `json:"server" yaml:"server"`
	Log                Log                      `json:"log" yaml:"log"`
	AppDataPointConfig AppDataPointConfigStruct `json:"dataPointConfig" yaml:"dataPointConfig"`
	MqttConfig         AppMqttConfigStruct      `json:"mqttConfig" yaml:"mqttConfig"`
	SystemInfo         AppSystemInfoStruct      `json:"systemInfo" yaml:"systemInfo"`
}
type Log struct {
	Level string `json:"level" yaml:"level"`
	Path  string `json:"path" yaml:"path"`
}
type AppDataPointConfigStruct struct {
	Path string `json:"path" yaml:"path"`
}

type AppMqttConfigStruct struct {
	Log struct {
		Enabled bool   `json:"enabled" yaml:"enabled"`
		Level   string `json:"level" yaml:"level"`
		File    string `json:"file" yaml:"file"`
	} `json:"log" yaml:"log"`
	File          string `json:"file" yaml:"file"`
	MessageConfig struct {
		Dir     string `json:"dir" yaml:"dir"`
		Publish struct {
			Path string `json:"path" yaml:"path"`
		} `json:"publish" yaml:"publish"`
		Subscribe struct {
			Path string `json:"path" yaml:"path"`
		} `json:"subscribe" yaml:"subscribe"`
	} `json:"messageConfig" yaml:"messageConfig"`
}
type AppSystemInfoStruct struct {
	File string `yaml:"file"`
}
type IAppConfigUseCase interface {
	ParseConfig()
	IsDebug() bool
	GetConfig() *AppConfig
	GetLogConfig() *Log
	GetAppDataPointConfig() *AppDataPointConfigStruct
	GetAppMqttConfig() *AppMqttConfigStruct
	GetAppSystemInfo() *AppSystemInfoStruct
}

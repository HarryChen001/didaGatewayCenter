package usecase

import (
	"didaGatewayCenter/domain"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var defaultConfig = domain.AppConfig{
	Debug: false,
	Server: struct {
		Address string `json:"address" yaml:"address"`
	}(struct {
		Address string `json:"address"`
	}{":8888"}),
	Log: domain.Log(struct {
		Level string `json:"level"`
		Path  string `json:"path"`
	}{"info", "/var/log/didaGatewayCenter/"}),
	AppDataPointConfig: domain.AppDataPointConfigStruct(struct {
		Path string `json:"path"`
	}{"/etc/didaGatewayCenter/config"}),
	MqttConfig: domain.AppMqttConfigStruct{
		Log: struct {
			Enabled bool   `json:"enabled" yaml:"enabled"`
			Level   string `json:"level" yaml:"level"`
			File    string `json:"file" yaml:"file"`
		}{},
		File: "",
		MessageConfig: struct {
			Dir     string `json:"dir" yaml:"dir"`
			Publish struct {
				Path string `json:"path" yaml:"path"`
			} `json:"publish" yaml:"publish"`
			Subscribe struct {
				Path string `json:"path" yaml:"path"`
			} `json:"subscribe" yaml:"subscribe"`
		}{},
	},
}

type AppConfigUseCase struct {
	config   *domain.AppConfig
	fileName string
}

func (n *AppConfigUseCase) IsDebug() bool {
	return n.config.Debug
}
func (n *AppConfigUseCase) GetAppSystemInfo() *domain.AppSystemInfoStruct {
	return &n.config.SystemInfo
}

func (n *AppConfigUseCase) GetAppMqttConfig() *domain.AppMqttConfigStruct {
	return &n.config.MqttConfig
}

func (n *AppConfigUseCase) GetAppDataPointConfig() *domain.AppDataPointConfigStruct {
	return &n.config.AppDataPointConfig
}

func (n *AppConfigUseCase) GetConfig() *domain.AppConfig {
	return n.config
}

func (n *AppConfigUseCase) GetLogConfig() *domain.Log {
	return &n.config.Log
}

func NewAppConfigUseCase(filename string) domain.IAppConfigUseCase {
	return &AppConfigUseCase{
		fileName: filename,
	}
}
func (n *AppConfigUseCase) ParseConfig() {

	config := &domain.AppConfig{}
	fileInfo, err := os.ReadFile(n.fileName)

	if os.IsNotExist(err) {
		config = &defaultConfig
		log.Println("Using default configuration file because the specify file is not exist:", n.fileName)
	} else if err == nil {
		if ok := json.Valid(fileInfo); !ok {
			if err := yaml.Unmarshal(fileInfo, config); err != nil {
				log.Println("Using default configuration file for error occurred while parse:", err)
				config = &defaultConfig
			}
			//			log.Println("Using the default configuration because the configuration file is not in json format")
		} else if err := json.Unmarshal(fileInfo, config); err != nil {
			config = &defaultConfig
			log.Println("Using the default configuration because an unknown error occurred while parsing the configuration file：", err.Error())
		}
	} else {
		config = &defaultConfig
		log.Println("Using the default configuration because there was an unknown error reading the file:", err.Error())
	}
	if config.Debug {
		log.Println("run in debug mode")
	}
	n.config = config
}

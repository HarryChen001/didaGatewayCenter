package domain

type System struct {
	MachineInfo MachineInfoStruct
}
type MachineInfoStruct struct {
	Name            string `yaml:"name"`
	Server          string `yaml:"server"`
	Port            int    `yaml:"port"`
	Sn              string `yaml:"sn"`
	Mac             string `yaml:"mac"`
	Enabled         bool   `yaml:"enabled"`
	UploadIntervalS int    `yaml:"UploadIntervalS"`
}

type ISystemUseCase interface {
	GetMachineInfoName() string
	GetMachineInfoSn() string
	GetMachineInfoMac() string
	GetMachineInfoServer() string
	GetMachineInfoPort() int
	GetMachineInfoEnabled() bool
	GetMachineInfoUploadIntervalS() int
}

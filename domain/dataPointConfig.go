package domain

import (
	"github.com/labstack/echo"
	"time"
)

type IDataPointConfigUseCase interface {
	GetPortConfigs() *Port
	GetPortConfigsByDeviceType(deviceType DeviceType) *Port
	GetDeviceConfigs(portName string) *DataPointDeviceConfig
	GetVariableConfigs(portName string) []*DataPointVariableConfig
}

type IDataPointConfigHandler interface {
	ConfigUpdate(ctx echo.Context) error
}

type Port struct {
	PortConfigs []DataPointPortConfig `json:"PORTConfigs"`
}

type DataPointPortConfig struct {
	PortName   string     `json:"PortName"`
	Vaild      bool       `json:"Vaild"`
	PortType   PortType   `json:"PortType"`
	DeviceType DeviceType `json:"DeviceType"`
	Param      PortParam  `json:"Param"`
}
type PortParam struct {
	COM      int    `json:"COM"`
	BandRate int    `json:"BandRate"`
	Parity   string `json:"Parity"`
	DateBits int    `json:"DateBits"`
	StopBit  int    `json:"StopBit"`

	ConvertPort   int  `json:"ConvertPort"`
	ConvertEnable bool `json:"ConvertEnable"`

	FrameIntervalMs int `json:"FrameIntervalMs"`

	IP         string `json:"IP"`
	PortNumber int    `json:"PortNumber"`

	RespTimeOutMs int `json:"RespTimeOutMs"`

	SampleIntervalS int `json:"SampleIntervalS"`
}

type DeviceType int

const (
	//Serial DeviceInfo Type
	DeviceTypeModbusRTU              DeviceType = 3001
	DeviceTypeModbusASCII            DeviceType = 3002
	DeviceTypeSerialPortPenetrate    DeviceType = 3003
	DeviceTypeDTL645                 DeviceType = 3006
	DeviceTypeMitsubishiProgramPort  DeviceType = 3007
	DeviceTypeMitsubishiComputerLink DeviceType = 3008
	DeviceTypeS7200PPI               DeviceType = 3009
	DeviceTypeHostLinkCMode          DeviceType = 3010
	DeviceTypeHostLinkFins1          DeviceType = 3011
	DeviceTypeHostLinkFins2          DeviceType = 3012

	DeviceTypeBridgeSerialPort DeviceType = 3014
	DeviceTypeInternal         DeviceType = 3013
	DeviceTypeInternal1        DeviceType = 3017

	//Net DeviceInfo Type
	DeviceTypeModbusTCP         DeviceType = 3005
	DeviceTypeHostLinkFinsTcp   DeviceType = 3101
	DeviceTypeSiemens200CP2431  DeviceType = 3102
	DeviceTypeSiemensS200Smart  DeviceType = 3103
	DeviceTypeSiemensS300       DeviceType = 3104
	DeviceTypeSiemensS400       DeviceType = 3105
	DeviceTypeSiemensS1200      DeviceType = 3106
	DeviceTypeSiemensS1500      DeviceType = 3107
	DeviceTypeSiemensFetchWrite DeviceType = 3108
	DeviceTypeMCBinaryQna3E     DeviceType = 3109
	DeviceTypeMCAsciiQna3E      DeviceType = 3110
	DeviceTypeMcBinaryQna1E     DeviceType = 3111
)

type PortType int

const (
	SerialType PortType = 1
	NetType    PortType = 2
)

type Device struct {
	DeviceConfigs []DataPointDeviceConfig `json:"DEVConfigs"`
}
type DataPointDeviceConfig struct {
	PortName string        `json:"PortName"`
	DevList  []*DeviceList `json:"DevList"`
}
type DeviceList struct {
	DevName       string    `json:"DevName"`
	OpcPath       string    `json:"OpcPath"`
	DevAddr       int       `json:"DevAddr"`
	FloatOrder    ByteOrder `json:"FloatOrder"`
	LongOrder     ByteOrder `json:"LongOrder"`
	LongLongOrder ByteOrder `json:"LongLongOrder"`
	DoubleOrder   ByteOrder `json:"DoubleOrder"`
}
type ByteOrder int

const (
	ByteOrderABCD ByteOrder = 1 + iota
	ByteOrderCDAB
	ByteOrderBADC
	ByteOrderDCBA
	//Long
	DeviceLongOrderABCD ByteOrder = 1
	DeviceLongOrderCDAB ByteOrder = 2
	DeviceLongOrderBADC ByteOrder = 3
	DeviceLongOrderDCBA ByteOrder = 4
	//Long Long
	DeviceLongLongOrderABCD ByteOrder = 1
	DeviceLongLongOrderCDAB ByteOrder = 2
	DeviceLongLongOrderBADC ByteOrder = 3
	DeviceLongLongOrderDCBA ByteOrder = 4
	//float32
	DeviceFloatOrderABCD ByteOrder = 1
	DeviceFloatOrderCDAB ByteOrder = 2
	DeviceFloatOrderBADC ByteOrder = 3
	DeviceFloatOrderDCBA ByteOrder = 4
	//float64
	DeviceDoubleOrderABCD ByteOrder = 1
	DeviceDoubleOrderCDAB ByteOrder = 2
	DeviceDoubleOrderBADC ByteOrder = 3
	DeviceDoubleOrderDCBA ByteOrder = 4
)

type Variable struct {
	VariableConfigs []DataPointVariableConfig `json:"VARConfigs"`
}

type DataPointVariableConfig struct {
	PortName string                  `json:"PortName"`
	DevName  string                  `json:"DevName"`
	VarList  []DataPointVariableList `json:"VarList"`
}

type DataPointVariableList struct {
	Id             int64    `json:"Id"`
	Name           string   `json:"Name"`
	AnotherName    string   `json:"AnotherName"`
	DataType       DataType `json:"DataType"`
	Decimal        int      `json:"Decimal"`
	Unit           string   `json:"Unit"`
	Modulus        float64  `json:"Modulus"`
	Offset         float64  `json:"Offset"`
	OpcVarPath     string   `json:"OpcVarPath"`
	SignalType     int      `json:"SignalType"`
	UpRangeValue   float64  `json:"UpRangeValue"`
	DownRangeValue float64  `json:"DownRangeValue"`
	Param          struct {
		DBNum   int          `json:"DBNum"`
		RegAddr int          `json:"RegAddr"`
		BitAddr int          `json:"BitAddr"`
		RegType RegisterType `json:"RegType"`
	} `json:"Param"`
	Event struct {
		EventName string `json:"EventName"`
		MathType  int    `json:"MathType"`
	} `json:"Event"`
	Value     interface{}
	Timestamp time.Time
}

type RegisterType int

const (
	// RegTypeCoilStatusWithWriteMultiple 01读线圈/15写线圈
	RegTypeCoilStatusWithWriteMultiple RegisterType = 1
	// RegTypeInputStatus 02读离散输入
	RegTypeInputStatus RegisterType = 2
	// RegTypeHoldingRegisterWithWriteMultiple 03读保持寄存器/16写保持寄存器
	RegTypeHoldingRegisterWithWriteMultiple RegisterType = 3
	// RegTypeInputRegister 04读输入寄存器
	RegTypeInputRegister RegisterType = 4
	// RegTypeCoilStatusWithWriteSingle 01读线圈/05写线圈
	RegTypeCoilStatusWithWriteSingle RegisterType = 5
	// RegTypeHoldingRegisterWithWriteSingle 03读保持寄存器/06写保持寄存器
	RegTypeHoldingRegisterWithWriteSingle RegisterType = 6
	//Mitsubishi Register
	RegTypeMitsubishiXRegister  RegisterType = 11 //Input
	RegTypeMitsubishiYRegister  RegisterType = 12 //Output
	RegTypeMitsubishiMRegister  RegisterType = 13 //Middle
	RegTypeMitsubishiSRegister  RegisterType = 14
	RegTypeMitsubishiTRegister  RegisterType = 15
	RegTypeMitsubishiCRegister  RegisterType = 16
	RegTypeMitsubishiDRegister  RegisterType = 17
	RegTypeMitsubishiTVRegister RegisterType = 18
	RegTypeMitsubishiCVRegister RegisterType = 19
	//Siemens Register
	RegTypeSiemensI  RegisterType = 21
	RegTypeSiemensQ  RegisterType = 22
	RegTypeSiemensM  RegisterType = 23
	RegTypeSiemensV  RegisterType = 24
	RegTypeSiemensSM RegisterType = 25
	RegTypeSiemensAI RegisterType = 26
	RegTypeSiemensAQ RegisterType = 27
	RegTypeSiemensT  RegisterType = 28
	RegTypeSiemensC  RegisterType = 29
	RegTypeSiemensDB RegisterType = 30
	//Omron Register
	RegTypeOmronCIORegister RegisterType = 31 //Input
	RegTypeOmronLRegister   RegisterType = 32 //Output
	RegTypeOmronHRegister   RegisterType = 33
	RegTypeOmronARegister   RegisterType = 34
	RegTypeOmronDMRegister  RegisterType = 35
	RegTypeOmronEMRegister  RegisterType = 36
	RegTypeOmronTSRegister  RegisterType = 37
	RegTypeOmronCSRegister  RegisterType = 38
	RegTypeOmronTVRegister  RegisterType = 39
	RegTypeOmronCVRegister  RegisterType = 40
	RegTypeOmronWARegister  RegisterType = 41
)

type DataType int

const (
	VarDataTypeBool   DataType = 1
	VarDataTypeUint16 DataType = 2
	VarDataTypeUint32 DataType = 3
	VarDataTypeUint64 DataType = 4
	VarDataTypeInt16  DataType = 5
	VarDataTypeInt32  DataType = 6
	VarDataTypeInt64  DataType = 7
	VarDataTypeFloat  DataType = 8
	VarDataTypeDouble DataType = 9
	VarDataTypeString DataType = 10
	VarDataTypeByte   DataType = 11
	VarDataTypeBit    DataType = 12
)

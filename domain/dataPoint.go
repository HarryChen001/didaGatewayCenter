package domain

import (
	"github.com/labstack/echo"
	"time"
)

type DataPoint struct {
	PortConfig     *DataPointPortConfig
	DeviceConfig   *DataPointDeviceConfig
	VariableConfig []*DataPointVariableConfig
	Driver         IDataPointDriverUsecase
}
type AllDataPoints struct {
	PortName     string      `json:"portName"`
	DeviceName   string      `json:"deviceName"`
	VariableName string      `json:"variableName"`
	Value        interface{} `json:"value"`
	Timestamp    time.Time   `json:"timestamp"`
}

type IDataPointUseCase interface {
	Read(portName string, deviceName string, variableName string, isRealTime bool) (interface{}, error)
	ReadById(id int64, isRealTime bool) (interface{}, error)
	WriteById(id int64, value interface{}) (interface{}, error)
	GetStore() []AllDataPoints
	CycleSample()
}
type IDataPointHandler interface {
	GetAllVariablesV1(ctx echo.Context) error
	GetAllVariablesV2(ctx echo.Context) error
}

package domain

import "time"

type DataPointDriver struct {
}

type IDataPointDriverUsecase interface {
	Init(portInfo *DataPointPortConfig, transform IDataTransformUsecase)
	Read(portInfo *DataPointPortConfig, deviceInfo *DeviceList, variableInfo *DataPointVariableList) IValueType
	Write(portInfo *DataPointPortConfig, deviceInfo *DeviceList, variableInfo *DataPointVariableList, value interface{}) error
}

type Software interface {
	ReadTimeout(t time.Duration) ([]byte, error)
	WriteTimeout(writeData []byte, t time.Duration) error
	WriteReadTimeout(writeData []byte, t time.Duration) ([]byte, error)
}

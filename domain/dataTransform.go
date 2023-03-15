package domain

type DataTransform struct {
}
type ValueType struct {
	input []byte
}
type IValueType interface {
	ToFloat64() float64
	OriginalValue() uint16
}

type IDataTransformUsecase interface {
	ByteToValue(list *DeviceList, variableList *DataPointVariableList, result []byte) (IValueType, error)
	ValueToByte(list *DeviceList, variableList *DataPointVariableList, input interface{}) ([]byte, error)
}

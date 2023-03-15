package dataTransform

import (
	"didaGatewayCenter/domain"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

type dataTransform struct {
}
type valueType struct {
	input  []byte
	output float64
}

func (d *valueType) ToFloat64() float64 {
	return d.output
}
func (d *valueType) OriginalValue() uint16 {
	return binary.BigEndian.Uint16(d.input)
}
func (d *dataTransform) ValueToByte(list *domain.DeviceList, variableList *domain.DataPointVariableList, input interface{}) ([]byte, error) {
	byteOrder := domain.ByteOrderABCD
	dataType := variableList.DataType
	modulus := variableList.Modulus
	offset := variableList.Offset

	var result []byte
	inputValue := input.(float64)
	if offset != 0 {
		inputValue -= offset
	}
	if modulus != 1 {
		inputValue /= modulus
	}
	switch dataType {
	case domain.VarDataTypeBit:
		result = make([]byte, 2)
		binary.BigEndian.PutUint16(result, uint16(input.(float64)))
		return result, nil
	case domain.VarDataTypeByte:
		result = make([]byte, 1)
		result[0] = byte(input.(float64))
		return result, nil
	case domain.VarDataTypeBool:
		if input.(float64) != 0 {
			return []byte{0x00, 0x01}, nil
		}
		return []byte{0x00, 0x00}, nil
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		result = make([]byte, 2)
		binary.BigEndian.PutUint16(result, uint16(inputValue))
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32:
		result = make([]byte, 4)
		byteOrder = list.LongOrder
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			binary.BigEndian.PutUint32(result, uint32(inputValue))
		} else {
			binary.LittleEndian.PutUint32(result, uint32(inputValue))
		}
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64:
		result = make([]byte, 8)
		byteOrder = list.LongLongOrder
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			binary.BigEndian.PutUint64(result, uint64(inputValue))
		} else {
			binary.LittleEndian.PutUint64(result, uint64(inputValue))
		}
	case domain.VarDataTypeFloat:
		result = make([]byte, 4)
		byteOrder = list.FloatOrder
		bits := math.Float32bits(float32(inputValue))
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			binary.BigEndian.PutUint32(result, bits)
		} else {
			binary.LittleEndian.PutUint32(result, bits)
		}
	case domain.VarDataTypeDouble:
		result = make([]byte, 8)
		byteOrder = list.DoubleOrder
		bits := math.Float64bits(inputValue)
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			binary.BigEndian.PutUint64(result, bits)
		} else {
			binary.LittleEndian.PutUint64(result, bits)
		}
	}
	if byteOrder == domain.ByteOrderBADC || byteOrder == domain.ByteOrderCDAB {
		swap(result)
	}
	return result, nil
}
func (d *dataTransform) ByteToValue(list *domain.DeviceList, variableList *domain.DataPointVariableList, result []byte) (domain.IValueType, error) {

	vType := &valueType{input: result}
	byteOrder := domain.ByteOrderABCD
	dataType := variableList.DataType
	length := 0
	switch dataType {
	case domain.VarDataTypeByte:
		length = 2
	case domain.VarDataTypeBit:
		length = 2
	case domain.VarDataTypeBool:
		length = 2
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		length = 2
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32:
		length = 4
		byteOrder = list.LongOrder
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64:
		length = 8
		byteOrder = list.LongLongOrder
	case domain.VarDataTypeFloat:
		length = 4
		byteOrder = list.FloatOrder
	case domain.VarDataTypeDouble:
		length = 8
		byteOrder = list.DoubleOrder
	}
	if len(result) != length {
		err := fmt.Errorf("%s the length of the input data %d is not equal to the require length %d,input %d", variableList.Name, len(result), length, result)
		return nil, err
	}
	if byteOrder == domain.ByteOrderCDAB || byteOrder == domain.ByteOrderBADC {
		swap(result)
	}
	var value interface{}
	switch dataType {
	case domain.VarDataTypeBool:
		vType.output = 0
		if result[1] != 0 {
			vType.output = 1
		}
		return vType, nil
	case domain.VarDataTypeByte:
		value = float64(result[1])
	case domain.VarDataTypeBit:
		switch variableList.Param.RegType {
		case domain.RegTypeSiemensI, domain.RegTypeSiemensQ, domain.RegTypeSiemensAQ, domain.RegTypeSiemensAI, domain.RegTypeSiemensSM,
			domain.RegTypeSiemensV, domain.RegTypeSiemensDB, domain.RegTypeSiemensM, domain.RegTypeSiemensC, domain.RegTypeSiemensT:
			vType.output = 0
			if result[1] != 0 {
				vType.output = 1
			}
			return vType, nil
		}
		bitAddress := variableList.Param.BitAddr
		r := int(result[0])*256 + int(result[1])
		//log.Println(r, result[0], result[1])
		v := (r >> bitAddress) & 0x01
		//		value = (result[len(result)-1-(bitAddress/8)] >> (bitAddress % 8)) & 0x01
		vType.output = float64(v)

		return vType, nil
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		value = binary.BigEndian.Uint16(result)
		if dataType == domain.VarDataTypeInt16 {
			value = float64(int16(value.(uint16)))
		} else {
			value = float64(value.(uint16))
		}
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32:
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			value = binary.BigEndian.Uint32(result)
		} else {
			value = binary.LittleEndian.Uint32(result)
		}
		if dataType == domain.VarDataTypeInt32 {
			value = float64(int32(value.(uint32)))
		} else {
			value = float64(value.(uint32))
		}
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64:
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			value = binary.BigEndian.Uint64(result)
		} else {
			value = binary.LittleEndian.Uint64(result)
		}
		if dataType == domain.VarDataTypeInt64 {
			value = float64(int64(value.(uint64)))
		} else {
			value = float64(value.(uint64))
		}
	case domain.VarDataTypeFloat:
		var bits uint32
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			bits = binary.BigEndian.Uint32(result)
		} else {
			bits = binary.LittleEndian.Uint32(result)
		}
		value = math.Float32frombits(bits)
	case domain.VarDataTypeDouble:
		var bits uint64
		if byteOrder == domain.ByteOrderABCD || byteOrder == domain.ByteOrderBADC {
			bits = binary.BigEndian.Uint64(result)
		} else {
			bits = binary.LittleEndian.Uint64(result)
		}
		value = math.Float64frombits(bits)
	}

	if variableList.Modulus != 1 || variableList.Offset != 0 {
		value = value.(float64)*variableList.Modulus + variableList.Offset
	}
	value, _ = strconv.ParseFloat(fmt.Sprintf("%."+strconv.Itoa(variableList.Decimal)+"f", value), 64)

	vType.output = value.(float64)
	return vType, nil
}
func swap(input []byte) {
	if len(input)%2 != 0 {
		return
	}
	for i := 0; i < len(input); i += 2 {
		input[i], input[i+1] = input[i+1], input[i]
	}
}
func NewDataTransformUsecase() domain.IDataTransformUsecase {
	d := dataTransform{}
	return &d
}

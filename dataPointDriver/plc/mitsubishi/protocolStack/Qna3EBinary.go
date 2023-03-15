package protocolStack

import (
	"encoding/binary"
	"fmt"
)

func (q *Qna3EBinaryProtocolStack) ReadVar(isBit bool, code *RegCode, startAddress int, length int) []byte {
	q.dataLength = 0x0c
	q.command = readBatch
	q.childCommand = childCommandWord
	if isBit {
		q.childCommand = childCommandBit
	}
	q.data1.code = *code
	q.data1.address = uint32(startAddress)
	q.data1.length = uint16(length)

	deputyHeader := make([]byte, 2)
	networkNum := q.networkNum
	plcNum := q.plcNum
	targetIONum := make([]byte, 2)
	targetModule := q.targetModuleStation
	dataLength := make([]byte, 2)
	cpuTimer := make([]byte, 2)
	command := make([]byte, 2)
	childCommand := make([]byte, 2)
	unitStartAddress := make([]byte, 4)
	unitCode := code.binaryCode
	unitLength := make([]byte, 2)

	binary.BigEndian.PutUint16(deputyHeader, q.deputyHeader)
	binary.LittleEndian.PutUint16(targetIONum, q.targetIONum)
	binary.LittleEndian.PutUint16(dataLength, q.dataLength)
	binary.LittleEndian.PutUint16(cpuTimer, q.cpuTimer)
	binary.LittleEndian.PutUint16(command, uint16(q.command))
	binary.LittleEndian.PutUint16(childCommand, uint16(q.childCommand))
	binary.LittleEndian.PutUint16(unitStartAddress, uint16(q.data1.address))
	binary.LittleEndian.PutUint16(unitLength, q.data1.length)

	result := append(append(deputyHeader, networkNum, plcNum), append(targetIONum, targetModule)...)
	result = append(result, append(dataLength, cpuTimer...)...)
	result = append(result, append(command, childCommand...)...)
	result = append(result, append(unitStartAddress[0:3], byte(unitCode))...)
	result = append(result, unitLength...)

	return result
}

func (q *Qna3EBinaryProtocolStack) Parse(input []byte) ([]byte, error) {
	if len(input) < 11 {
		return nil, fmt.Errorf("invalid length: %d", len(input))
	}
	endCode := binary.LittleEndian.Uint16(input[9:11])
	dataLength := binary.LittleEndian.Uint16(input[7:9])
	if endCode != 0 {
		return nil, fmt.Errorf("invalid end code: %d", endCode)
	}
	if dataLength == 2 {
		return nil, nil
	}
	result := make([]byte, len(input[11:]))
	for index, singleR := range input[11:] {
		result[len(input[11:])-1-index] = singleR
	}
	return result, nil
}

func (q *Qna3EBinaryProtocolStack) WriteVar(isBit bool, code *RegCode, startAddress int, length int, data []byte) []byte {
	q.dataLength = uint16(0x0c + len(data))
	q.command = writeBatch
	q.childCommand = childCommandWord
	if isBit {
		q.childCommand = childCommandBit
		if data[0] == 1 {
			data[0] = 0x10
		}
	}
	q.data1.code = *code
	q.data1.address = uint32(startAddress)
	q.data1.length = uint16(length)

	deputyHeader := make([]byte, 2)
	networkNum := q.networkNum
	plcNum := q.plcNum
	targetIONum := make([]byte, 2)
	targetModule := q.targetModuleStation
	dataLength := make([]byte, 2)
	cpuTimer := make([]byte, 2)
	command := make([]byte, 2)
	childCommand := make([]byte, 2)
	unitStartAddress := make([]byte, 4)
	unitCode := code.binaryCode
	unitLength := make([]byte, 2)

	binary.BigEndian.PutUint16(deputyHeader, q.deputyHeader)
	binary.LittleEndian.PutUint16(targetIONum, q.targetIONum)
	binary.LittleEndian.PutUint16(dataLength, q.dataLength)
	binary.LittleEndian.PutUint16(cpuTimer, q.cpuTimer)
	binary.LittleEndian.PutUint16(command, uint16(q.command))
	binary.LittleEndian.PutUint16(childCommand, uint16(q.childCommand))
	binary.LittleEndian.PutUint16(unitStartAddress, uint16(q.data1.address))
	binary.LittleEndian.PutUint16(unitLength, q.data1.length)

	result := append(append(deputyHeader, networkNum, plcNum), append(targetIONum, targetModule)...)
	result = append(result, append(dataLength, cpuTimer...)...)
	result = append(result, append(command, childCommand...)...)
	result = append(result, append(unitStartAddress[0:3], byte(unitCode))...)
	result = append(result, unitLength...)

	writeData := make([]byte, len(data))
	for index, singleD := range data {
		writeData[len(data)-index-1] = singleD
	}
	result = append(result, writeData...)

	return result
}

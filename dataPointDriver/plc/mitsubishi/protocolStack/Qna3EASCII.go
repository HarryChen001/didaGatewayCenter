package protocolStack

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func (q *Qna3EAsciiProtocolStack) ReadVar(isBit bool, code *RegCode, startAddress int, length int) []byte {
	q.dataLength = 0x18
	q.command = readBatch
	q.childCommand = childCommandWord
	if isBit {
		q.childCommand = childCommandBit
	}
	q.data1.code = *code
	q.data1.address = uint32(startAddress)
	q.data1.length = uint16(length)

	deputyHeaderAscii := fmt.Sprintf("%.4X", q.deputyHeader)
	networkNumAscii := fmt.Sprintf("%.2X", q.networkNum)
	plcNum := fmt.Sprintf("%.2X", q.plcNum)
	targetIONumAscii := fmt.Sprintf("%.4X", q.targetIONum)
	targetModuleStationAscii := fmt.Sprintf("%.2X", q.targetModuleStation)
	dataLength := fmt.Sprintf("%.4X", q.dataLength)
	cpuTimer := fmt.Sprintf("%.4X", q.cpuTimer)
	commandASCII := fmt.Sprintf("%.4X", q.command)
	childCommandASCII := fmt.Sprintf("%.4X", q.childCommand)
	uintCode := q.data1.code.asciiCode
	startUint := fmt.Sprintf("%.6d", q.data1.address)
	uintLength := fmt.Sprintf("%.4X", q.data1.length)

	result := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s", deputyHeaderAscii, networkNumAscii, plcNum, targetIONumAscii, targetModuleStationAscii,
		dataLength, cpuTimer, commandASCII, childCommandASCII, uintCode, startUint, uintLength)
	return []byte(result)
}

func (q *Qna3EAsciiProtocolStack) Parse(input []byte) ([]byte, error) {
	if len(input) < 22 {
		return nil, fmt.Errorf("invalid length: %d", len(input))
	}
	endCodeByte, _ := hex.DecodeString(string(input[18:22]))
	endCode := binary.BigEndian.Uint16(endCodeByte)
	dataLengthByte, _ := hex.DecodeString(string(input[14:18]))
	dataLength := binary.BigEndian.Uint16(dataLengthByte)
	if endCode != 0 {
		return nil, fmt.Errorf("invalid end code: %d", endCode)
	}
	if dataLength == 2 {
		return nil, nil
	}
	result := make([]byte, len(input[22:]))
	if len(result) == 1 {
		if result[0] == 0x31 {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	}
	r, _ := hex.DecodeString(string(input[22:]))

	return r, nil
}

func (q *Qna3EAsciiProtocolStack) WriteVar(isBit bool, code *RegCode, startAddress int, length int, data []byte) []byte {
	q.dataLength = uint16(0x18 + len(data)*2)
	q.command = writeBatch
	if isBit {
		q.childCommand = childCommandBit
		if data[0] == 1 {
			data[0] = 0x10
		}
	}
	q.data1.code = *code
	q.data1.address = uint32(startAddress)
	q.data1.length = uint16(length)

	deputyHeaderAscii := fmt.Sprintf("%.4X", q.deputyHeader)
	networkNumAscii := fmt.Sprintf("%.2X", q.networkNum)
	plcNum := fmt.Sprintf("%.2X", q.plcNum)
	targetIONumAscii := fmt.Sprintf("%.4X", q.targetIONum)
	targetModuleStationAscii := fmt.Sprintf("%.2X", q.targetModuleStation)
	dataLength := fmt.Sprintf("%.4X", q.dataLength)
	cpuTimer := fmt.Sprintf("%.4X", q.cpuTimer)
	commandASCII := fmt.Sprintf("%.4X", q.command)
	childCommandASCII := fmt.Sprintf("%.4X", q.childCommand)
	uintCode := q.data1.code.asciiCode
	startUint := fmt.Sprintf("%.6d", q.data1.address)
	uintLength := fmt.Sprintf("%.4X", q.data1.length)

	for i := 0; i < len(data); i += 2 {
		data[i], data[len(data)-1-i] = data[len(data)-1-i], data[i]
	}
	for i := 0; i < len(data); i += 2 {
		data[i], data[i+1] = data[i+1], data[i]
	}
	writeDataByte := fmt.Sprintf("%X", data)
	if isBit {
		writeDataByte = string(writeDataByte[0])
	}

	result := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s", deputyHeaderAscii, networkNumAscii, plcNum, targetIONumAscii, targetModuleStationAscii,
		dataLength, cpuTimer, commandASCII, childCommandASCII, uintCode, startUint, uintLength, writeDataByte)
	return []byte(result)
}

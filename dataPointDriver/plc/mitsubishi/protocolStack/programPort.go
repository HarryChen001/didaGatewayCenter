package protocolStack

import (
	"encoding/hex"
	"fmt"
	"log"
)

const (
	stx = 0x02
)

func (p ProgramPort) ReadVar(isBit bool, code *RegCode, startAddress int, length int) []byte {

	address := code.binaryCode

	switch code {
	case &SerialInput, &SerialOutput:
		address += uint16(startAddress / 10)
	case &SerialData, &SerialCounter16Value:
		address += uint16(startAddress * 2)
		length *= 2
	case &SerialCounter32Value:
		address += uint16((startAddress - 200) * 4)
		length *= 2
	default:
		address += uint16(startAddress / 8)
	}
	addressByte := fmt.Sprintf("%.4X", address)
	result := []byte{0x02, 0x30}
	result = append(result, fmt.Sprintf("%s%s", addressByte, fmt.Sprintf("%.2X", length))...)
	result = append(result, 0x03)
	checkSum := 0
	for _, temp := range result[1:] {
		checkSum += int(temp)
		checkSum = int(checkSum) % 256
	}
	strCheckSum := fmt.Sprintf("%.2X", checkSum)
	result = append(result, strCheckSum...)
	log.Printf("% X,%s,%X", result, code.asciiCode, address)
	return result
}

func (p ProgramPort) WriteVar(isBit bool, code *RegCode, startAddress int, length int, data []byte) []byte {
	address := code.binaryCode

	switch code {
	case &SerialInput, &SerialOutput:
		address += uint16(startAddress / 10)
		data = []byte{data[1]}
	case &SerialData, &SerialCounter16Value:
		address += uint16(startAddress * 2)
		length *= 2
	case &SerialCounter32Value:
		address += uint16(startAddress * 4)
		length *= 2
	case &SerialSpecialRelay:
		address += uint16(startAddress / 8)
		data = []byte{data[1]}
	default:
		address += uint16(startAddress / 8)
	}

	if isBit {
		addressByte := []byte{0, 0, 0, 0}
		switch code {
		case &SerialStatus:
			address = uint16(startAddress) + 0x0000
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialInput:
			address = uint16(startAddress) + 0x0400
			addressByte = []byte(fmt.Sprintf("%.4d", address))
		case &SerialOutput:
			address = uint16(startAddress) + 0x0500
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialSpecialRelay:
			address = uint16(startAddress) + 0x0800
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialCounterCoil:
			address = uint16(startAddress) + 0x03c0
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialCounter16Value, &SerialCounter32Value:
			address = uint16(startAddress) + 0x0e00
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialTimerCoil:
			address = uint16(startAddress) + 0x02c0
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		case &SerialTimerValue:
			address = uint16(startAddress) + 0x0600
			addressByte = []byte(fmt.Sprintf("%.4X", address))
		}
		v := 0x37
		if data[0] == 0 {
			v = 0x38
		}
		command := []byte{stx, byte(v)}
		command = append(command, addressByte...)
		command = append(command, 0x03)
		checkSum := fxCheckSum(command[1:])
		checkSumByte := fmt.Sprintf("%.2X", checkSum)
		command = append(command, checkSumByte...)
		log.Printf("% X", command)
		return command
	}

	addressByte := fmt.Sprintf("%.4X", address)
	result := []byte{0x02, 0x31}
	result = append(result, fmt.Sprintf("%s%s", addressByte, fmt.Sprintf("%.2X", length))...)

	a := fmt.Sprintf("%X", data)
	result = append(result, a...)
	result = append(result, 0x03)
	checkSum := 0
	for _, temp := range result[1:] {
		checkSum += int(temp)
		checkSum = int(checkSum) % 256
	}
	strCheckSum := fmt.Sprintf("%.2X", checkSum)
	result = append(result, strCheckSum...)
	log.Printf("% X", result)
	return result
}

func (p ProgramPort) Parse(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("no any data")
	}
	switch input[0] {
	case 0x06:
		return nil, nil
	case 0x15:
		return nil, fmt.Errorf("error ack")
	case 0x02:
		goto p
	default:
		return nil, fmt.Errorf("unknown error:% X", input[0])
	}
p:
	if len(input) < 5 {
		return nil, fmt.Errorf("length is %d but expected at least 5", len(input))
	}
	checkSum := 0
	for _, temp := range input[1 : len(input)-2] {
		checkSum += int(temp)
		checkSum = int(checkSum) % 256
	}
	strCheckSum := fmt.Sprintf("%.2X", checkSum)
	sum := input[len(input)-2:]
	if strCheckSum != string(sum) {
		return nil, fmt.Errorf("checksum is %s not mactch % X which received", strCheckSum, sum)
	}

	v, _ := hex.DecodeString(string(input[1 : len(input)-3]))
	r := make([]byte, len(v))
	for index, single := range v {
		r[len(v)-1-index] = single
	}
	return r, nil
}
func fxCheckSum(input []byte) byte {
	checkSum := 0
	for _, temp := range input {
		checkSum += int(temp)
		checkSum = int(checkSum) % 256
	}
	return byte(checkSum)
}

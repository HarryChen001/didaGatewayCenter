package protocolStack

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const tPKTLength = 4

type plcType string
type MsgType byte
type FunctionCode byte
type Area byte
type TransportSize byte
type SyntaxId byte
type TransportSizeInData byte

type ReturnCode map[byte]string

var (
	returnDataCode ReturnCode = ReturnCode{
		0x00: "reserved",
		0x01: "Hardware error",
		0x03: "Accessing the object not allowed",
		0x05: "Invalid address",
		0x06: "Data type not supported",
		0x07: "Data type inconsistent",
		0x0a: "Object does not exist",
		0xff: "Success",
	}
)

const (
	S1500     plcType = "S1500"
	S1200     plcType = "S1200"
	S400      plcType = "S400"
	S300      plcType = "S300"
	S200      plcType = "S200"
	S200Smart plcType = "S200Smart"
)
const (
	MsgTypeJobRequest MsgType = 0x01
	MsgTypeAck        MsgType = 0x02
	MsgTypeAckData    MsgType = 0x03
	MsgTypeUserData   MsgType = 0x07
)
const (
	AreaTypeDATARecord Area = 0x01 /* Data record, used with RDREC or firmware updates on CP */
	AreaTypeSYSInfo    Area = 0x03 /* System info of 200 family */
	AreaTypeSYSFlags   Area = 0x05 /* System flags of 200 family */
	AreaTypeANAIN      Area = 0x06 /* analog inputs of 200 family */
	AreaTypeANAOUT     Area = 0x07 /* analog outputs of 200 family */
	AreaTypeCOUNTER    Area = 0x1C /* S7 counters */
	AreaTypeTIMER      Area = 0x1D /* S7 timers */
	AreaTypeCOUNTER200 Area = 0x1E /* IEC counters (200 family) */
	AreaTypeTIMER200   Area = 0x1F /* IEC timers (200 family) */
	AreaTypeP          Area = 0x80 /* direct peripheral access */
	AreaTypeINPUTS     Area = 0x81
	AreaTypeOUTPUTS    Area = 0x82
	AreaTypeFLAGS      Area = 0x83
	AreaTypeDB         Area = 0x84 /* data blocks */
	AreaTypeDI         Area = 0x85 /* instance data blocks */
	AreaTypeLOCAL      Area = 0x86 /* local data (should not be accessible over network) */
	AreaTypeV          Area = 0x87 /* previous (Vorgaenger) local data (should not be accessible over network)  */
)
const (
	// TransportSizeBit 1 byte
	TransportSizeBit TransportSize = 1 + iota
	// TransportSizeByte 1 byte
	TransportSizeByte
	// Char 1 byte
	Char
	// Word 2 byte
	Word
	// Int 2 byte
	Int
	// DWord 4 byte
	DWord
	// Dint 4 Byte
	Dint
	// Real special byte
	Real
	// Date special byte
	Date
	// Tod special byte
	Tod
	// Time special byte
	Time
	// S5Time special byte
	S5Time
	// Dt special byte
	Dt
	Counter
	Timer
	IECCounter
	IECTimer
	HSCounter
)

const (
	SyntaxIdS7Any            SyntaxId = 0x10 /* Address data S7-Any pointer-like DB1.DBX10.2 */
	SyntaxIdShort            SyntaxId = 0x11
	SyntaxIdExt              SyntaxId = 0x12
	SyntaxIdPbcId            SyntaxId = 0x13 /* R_ID for PBC */
	SyntaxIdAlarmLockFreeSet SyntaxId = 0x15 /* Alarm lock/free dataset */
	SyntaxIdAlarmIndSet      SyntaxId = 0x16 /* Alarm indication dataset */
	SyntaxIdAlarmAckSet      SyntaxId = 0x19 /* Alarm acknowledge message dataset */
	SyntaxIdAlarmQueryReqSet SyntaxId = 0x1a /* Alarm query request dataset */
	SyntaxIdNotifyIndSet     SyntaxId = 0x1c /* Notify indication dataset */
	SyntaxIdNCK              SyntaxId = 0x82 /* Sinumerik NCK HMI access (current units) */
	SyntaxIdNCKMetric        SyntaxId = 0x83 /* Sinumerik NCK HMI access metric units */
	SyntaxIdNCKInch          SyntaxId = 0x84 /* Sinumerik NCK HMI access inch */
	SyntaxIdDriveESAny       SyntaxId = 0xa2 /* seen on Drive ES Starter with routing over S7 */
	SyntaxId1200Symbolic     SyntaxId = 0xb2 /* Symbolic address mode of S7-1200 */
	SyntaxIdDBRead           SyntaxId = 0xb0 /* Kind of DB block read, seen only at an S7-400 */
)

const (
	DataTransportSizeNULL    TransportSizeInData = 0
	DataTransportSizeBBit    TransportSizeInData = 3  /* bit access, len is in bits */
	DataTransportSizeByte    TransportSizeInData = 4  /* byte/word/dword access, len is in bits */
	DataTransportSizeBInt    TransportSizeInData = 5  /* integer access, len is in bits */
	DataTransportSizeDint    TransportSizeInData = 6  /* integer access, len is in bytes */
	DataTransportSizeReal    TransportSizeInData = 7  /* real access, len is in bytes */
	DataTransportSizeBStr    TransportSizeInData = 9  /* octet string, len is in bytes */
	DataTransportSizeNCKDDR1 TransportSizeInData = 17 /* NCK address description, fixed length */
	DataTransportSizeNCKDDR2 TransportSizeInData = 18 /* NCK address description, fixed length */
)

const (
	FuncCodeCPUServices        FunctionCode = 0x00
	FuncCodeModeTransition     FunctionCode = 0x01
	FuncCodeSetupCommunication FunctionCode = 0xF0
	FuncCodeReadVar            FunctionCode = 0x04
	FuncCodeWriteVar           FunctionCode = 0x05
	FuncCodeRequestDownload    FunctionCode = 0x1A
	FuncCodeDownloadBlock      FunctionCode = 0x1B
	FuncCodeDownloadEnded      FunctionCode = 0x1C
	FuncCodeStartUpload        FunctionCode = 0x1D
	FuncCodeUpload             FunctionCode = 0x1E
	FuncCodeEndUpload          FunctionCode = 0x1F
	FuncCodePIService          FunctionCode = 0x28
	FuncCodePLCStop            FunctionCode = 0x29
)

var (
	S1200ParamMeter = []CotpParameter{{TPDUSize, 1, []byte{0x0a}},
		{SrcTSAP, 2, []byte{0x01, 0x01}},
		{DstTSAP, 2, []byte{0x01, 0x00}}}
	S1500ParamMeter = []CotpParameter{{TPDUSize, 1, []byte{0x0a}},
		{SrcTSAP, 2, []byte{0x01, 0x02}},
		{DstTSAP, 2, []byte{0x01, 0x00}}}
	S300ParamMeter = []CotpParameter{{TPDUSize, 1, []byte{0x0a}},
		{SrcTSAP, 2, []byte{0x01, 0x02}},
		{DstTSAP, 2, []byte{0x01, 0x00}}}
	S400ParamMeter = []CotpParameter{{TPDUSize, 1, []byte{0x0a}},
		{SrcTSAP, 2, []byte{0x01, 0x00}},
		{DstTSAP, 2, []byte{0x01, 0x00}}}
	S200ParamMeter = []CotpParameter{{SrcTSAP, 2, []byte{'M', 'W'}},
		{DstTSAP, 2, []byte{'M', 'W'}},
		{TPDUSize, 1, []byte{0x00}}}
	S200SmartParamMeter = []CotpParameter{{SrcTSAP, 2, []byte{0x10, 0x00}},
		{DstTSAP, 2, []byte{0x03, 0x00}},
		{TPDUSize, 1, []byte{0x00}}}
)

type s7CommHeader struct {
	protocolId    byte
	msgType       MsgType
	reserved1     byte
	reserved2     byte
	pduRef1       byte
	pduRef2       byte
	paramsLength1 byte
	paramsLength2 byte
	dataLength1   byte
	dataLength2   byte
}
type s7CommItem struct {
	varSpec       byte // variable specification
	addressLength byte
	syntaxId      SyntaxId      // syntax identifier
	transportSize TransportSize //
	length1       byte
	length2       byte
	dbNum1        byte
	dbNum2        byte
	area          Area
	address       []byte
}
type s7CommParameter struct {
	funcCode       FunctionCode
	reserved       byte
	maxAMQCalling1 byte
	maxAMQCalling2 byte
	maxAMQCalled1  byte
	maxAMQCalled2  byte
	pduLength1     byte
	pduLength2     byte
	itemCount      byte
	item           []s7CommItem
}
type s7Data struct {
	returnCode    byte
	transportSize TransportSizeInData
	length1       byte
	length2       byte
	data          []byte
}
type S7Comm struct {
	rack      byte // 机架号
	slot      byte // 槽位号
	tpkt      *Tpkt
	cotp      *CoTP
	random    *rand.Rand
	header    s7CommHeader
	parameter s7CommParameter
	data      s7Data
}

func NewS7Comm(rack byte, slot byte) *S7Comm {
	s := S7Comm{rack: rack, slot: slot, tpkt: NewTPKT(), cotp: NewCoTPUsecase()}
	s.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &s
}
func (s *S7Comm) GetCoTPShakeHands(plcType2 plcType) []byte {
	c := s.cotp
	c.SetPduType(ConnectRequestCR)
	switch plcType2 {
	case S1500:
		c.SetParameter(S1500ParamMeter)
	case S1200:
		c.SetParameter(S1200ParamMeter)
	case S400:
		c.SetParameter(S400ParamMeter)
	case S300:
		c.SetParameter(S300ParamMeter)
	case S200:
		c.SetParameter(S200ParamMeter)
	case S200Smart:
		c.SetParameter(S200SmartParamMeter)
	}
	c.param[len(c.param)-1].Data[len(c.param[len(c.param)-1].Data)-1] = s.rack*0x20 + s.slot
	cB := c.Byte()
	tB := NewTPKT().SetLength(c.Length() + 5).SetVersion(3).Byte()

	shakeHands := append(tB, cB...)
	return shakeHands
}

func (s *S7Comm) headerByte() []byte {
	h := s.header
	return []byte{h.protocolId, byte(h.msgType), h.reserved1, h.reserved2,
		h.pduRef1, h.pduRef2, h.paramsLength1, h.paramsLength2, h.dataLength1, h.dataLength2}
}
func (s *S7Comm) GetCommunicationByte() []byte {
	s.cotp.SetPduType(DataDT)
	pduRef := rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(65536)
	s.header = s7CommHeader{
		protocolId:    0x32,
		msgType:       MsgTypeJobRequest,
		reserved1:     0,
		reserved2:     0,
		pduRef1:       byte((pduRef & 0xff00) >> 8),
		pduRef2:       byte(pduRef & 0x00ff),
		paramsLength1: 0,
		paramsLength2: 8,
		dataLength1:   0,
		dataLength2:   0,
	}
	s.parameter = s7CommParameter{
		funcCode:       FuncCodeSetupCommunication,
		reserved:       0,
		maxAMQCalling1: 0,
		maxAMQCalling2: 1,
		maxAMQCalled1:  0,
		maxAMQCalled2:  1,
		pduLength1:     0x03,
		pduLength2:     0xc0,
	}
	s.tpkt.SetVersion(3)
	s.tpkt.SetLength(25)
	var result []byte
	result = append(result, s.tpkt.Byte()...)
	result = append(result, s.cotp.Byte()...)
	result = append(result, s.headerByte()...)
	result = append(result, byte(s.parameter.funcCode), s.parameter.reserved,
		s.parameter.maxAMQCalling1, s.parameter.maxAMQCalling2,
		s.parameter.maxAMQCalled1, s.parameter.maxAMQCalled2,
		s.parameter.pduLength1, s.parameter.pduLength2)
	return result
}

func (s *S7Comm) ReadVar(size TransportSize, sizeCount int, dbNum int, area Area, address1 int, address2 int) []byte {
	r := s.random.Int31n(65535)
	s.header = s7CommHeader{
		protocolId: 0x32,
		msgType:    MsgTypeJobRequest,
		reserved1:  0x00, reserved2: 0x00,
		pduRef1: byte(r >> 8), pduRef2: byte(r & 0x00ff),
		paramsLength1: 0, paramsLength2: 0, dataLength1: 0, dataLength2: 0,
	}
	address := (address1 << 3) | address2
	bA := make([]byte, 4)
	binary.BigEndian.PutUint32(bA, uint32(address))
	s.parameter = s7CommParameter{
		funcCode:  FuncCodeReadVar,
		itemCount: 0,
		item: []s7CommItem{{
			varSpec:       0x12,
			syntaxId:      SyntaxIdS7Any,
			transportSize: size,
			length1:       byte(sizeCount / 256), length2: byte(sizeCount % 256),
			dbNum1: byte(dbNum >> 8), dbNum2: byte(dbNum & 0x00ff),
			area:    area,
			address: []byte{bA[1], bA[2], bA[3]},
		}},
	}
	s.parameter.itemCount = byte(len(s.parameter.item))
	paramsByte := []byte{byte(s.parameter.funcCode), s.parameter.itemCount}
	for _, singleItem := range s.parameter.item {
		variableByte := []byte{byte(singleItem.syntaxId), byte(singleItem.transportSize), singleItem.length1, singleItem.length2,
			singleItem.dbNum1, singleItem.dbNum2, byte(singleItem.area)}
		variableByte = append(variableByte, singleItem.address...)
		paramsByte = append(paramsByte, singleItem.varSpec, byte(len(variableByte)))
		paramsByte = append(paramsByte, variableByte...)
	}
	s.header.paramsLength1 = byte(len(paramsByte) / 256)
	s.header.paramsLength2 = byte(len(paramsByte) % 256)
	headerByte := s.headerByte()
	cotpByte := s.cotp.SetPduType(DataDT).Byte()
	tpktByte := s.tpkt.SetVersion(0x03).Byte()
	result := tpktByte
	result = append(result, cotpByte...)
	result = append(result, headerByte...)
	result = append(result, paramsByte...)
	result[2] = byte(len(result) / 256)
	result[3] = byte(len(result) % 256)
	return result
}
func (s *S7Comm) Parse(i []byte) ([]byte, error) {
	if len(i) < 4 {
		return nil, fmt.Errorf("invalid length of bytes")
	}
	length := binary.BigEndian.Uint16(i[2:4])
	if len(i) != int(length) {
		return nil, errors.New(fmt.Sprintf("incomplete data,require %d but got %d", length, len(i)))
	}
	cotpLength := i[4]
	s7CommTotal := i[tPKTLength+cotpLength+1:]
	s7Header := s7CommTotal[:12]
	parameterLength := binary.BigEndian.Uint16(s7Header[6:8])
	s7DataLength := binary.BigEndian.Uint16(s7Header[8:10])
	s7Parameter := s7CommTotal[len(s7Header) : len(s7Header)+int(parameterLength)]
	s7CommData := s7CommTotal[len(s7Header)+len(s7Parameter):]
	errorClass := s7Header[10]
	errorCode := s7Header[11]
	if errorClass != 0 || errorCode != 0 {
		return nil, errors.New("error class or error code")
	}
	if s7DataLength == 0 {
		return nil, errors.New("data length is 0")
	}
	dataCode := s7CommData[0]
	if dataCode != 0xff {
		if returnDataString, ok := returnDataCode[dataCode]; ok {
			return nil, fmt.Errorf("%s(%X)", returnDataString, dataCode)
		} else {
			return nil, fmt.Errorf("%s(%X)", "unkown error", dataCode)
		}
	}
	if s7DataLength >= 5 {
		return s7CommData[4:], nil
	}
	return nil, nil
}
func (s *S7Comm) WriteVar(size TransportSize, sizeCount int, dbNum int, area Area, address1 int, address2 int, data []byte) []byte {
	r := s.random.Int31n(65535)
	s.header = s7CommHeader{
		protocolId: 0x32,
		msgType:    MsgTypeJobRequest,
		reserved1:  0x00, reserved2: 0x00,
		pduRef1: byte(r >> 8), pduRef2: byte(r & 0x00ff),
		paramsLength1: 0, paramsLength2: 0, dataLength1: 0, dataLength2: 0,
	}
	address := (address1 << 3) | address2
	bA := make([]byte, 4)
	binary.BigEndian.PutUint32(bA, uint32(address))
	s.parameter = s7CommParameter{
		funcCode:  FuncCodeWriteVar,
		itemCount: 0,
		item: []s7CommItem{{
			varSpec:       0x12,
			syntaxId:      SyntaxIdS7Any,
			transportSize: size,
			length1:       byte(sizeCount / 256), length2: byte(sizeCount % 256),
			dbNum1: byte(dbNum >> 8), dbNum2: byte(dbNum & 0x00ff),
			area:    area,
			address: []byte{bA[1], bA[2], bA[3]},
		}},
	}
	s.data = s7Data{
		returnCode: 0x00,
		data:       data,
	}
	if size == TransportSizeBit {
		s.data.transportSize = DataTransportSizeBBit
		s.data.length1 = byte(len(data) / 256)
		s.data.length2 = byte(len(data) % 256)
	} else {
		s.data.transportSize = DataTransportSizeBInt
		s.data.length1 = byte(len(data) * 8 / 256)
		s.data.length2 = byte(len(data) * 8 % 256)
	}
	dataByte := []byte{s.data.returnCode, byte(s.data.transportSize), s.data.length1, s.data.length2}
	dataByte = append(dataByte, s.data.data...)
	s.header.dataLength1 = byte(len(dataByte) / 256)
	s.header.dataLength2 = byte(len(dataByte) % 256)
	s.parameter.itemCount = byte(len(s.parameter.item))
	paramsByte := []byte{byte(s.parameter.funcCode), s.parameter.itemCount}
	for _, singleItem := range s.parameter.item {
		variableByte := []byte{byte(singleItem.syntaxId), byte(singleItem.transportSize), singleItem.length1, singleItem.length2,
			singleItem.dbNum1, singleItem.dbNum2, byte(singleItem.area)}
		variableByte = append(variableByte, singleItem.address...)
		paramsByte = append(paramsByte, singleItem.varSpec, byte(len(variableByte)))
		paramsByte = append(paramsByte, variableByte...)
	}
	s.header.paramsLength1 = byte(len(paramsByte) / 256)
	s.header.paramsLength2 = byte(len(paramsByte) % 256)
	headerByte := s.headerByte()
	cotpByte := s.cotp.SetPduType(DataDT).Byte()
	tpktByte := s.tpkt.SetVersion(0x03).Byte()
	result := tpktByte
	result = append(result, cotpByte...)
	result = append(result, headerByte...)
	result = append(result, paramsByte...)
	result = append(result, dataByte...)
	result[2] = byte(len(result) / 256)
	result[3] = byte(len(result) % 256)
	return result
}

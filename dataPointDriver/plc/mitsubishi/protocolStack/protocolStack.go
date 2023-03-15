package protocolStack

type commandType uint16
type childCommandType int
type RegCode struct {
	binaryCode uint16
	asciiCode  []byte
}

var (
	SpecialRelay            = RegCode{0x91, []byte{'S', 'M'}}
	SpecialRegister         = RegCode{0xA9, []byte{'S', 'D'}}
	Input                   = RegCode{0x9C, []byte{'X', '*'}}
	Output                  = RegCode{0x9D, []byte{'Y', '*'}}
	InternalRelay           = RegCode{0x90, []byte{'M', '*'}}
	LatchRelay              = RegCode{0x92, []byte{'L', '*'}}
	Alarm                   = RegCode{0x93, []byte{'F', '*'}}
	VariableAddress         = RegCode{0x94, []byte{'V', '*'}}
	LinkRelay               = RegCode{0xA0, []byte{'B', '*'}}
	DataRegister            = RegCode{0xA8, []byte{'D', '*'}}
	LinkRegister            = RegCode{0xB4, []byte{'W', '*'}}
	TimerContact            = RegCode{0xC1, []byte{'T', 'S'}} //定时器触点
	TimerCoil               = RegCode{0xC0, []byte{'T', 'C'}} //定时器线圈
	TimerCurrent            = RegCode{0xC2, []byte{'T', 'N'}} //定时器当前值
	SumTimerContact         = RegCode{0xC7, []byte{'S', 'S'}}
	SumTimerCoil            = RegCode{0xC6, []byte{'S', 'C'}}
	SumTimerCurrent         = RegCode{0xC8, []byte{'S', 'N'}}
	CounterContact          = RegCode{0xC4, []byte{'C', 'S'}}
	CounterCoil             = RegCode{0xC3, []byte{'C', 'C'}}
	CounterCurrent          = RegCode{0xC5, []byte{'C', 'N'}}
	LinkSpecialRelay        = RegCode{0xA1, []byte{'S', 'B'}}
	LinkSpecialRegister     = RegCode{0xB5, []byte{'S', 'W'}}
	StepRelay               = RegCode{0x98, []byte{'S', '*'}}
	DirectInput             = RegCode{0xA2, []byte{'D', 'X'}}
	DirectOutput            = RegCode{0xA3, []byte{'D', 'Y'}}
	VariableAddressRegister = RegCode{0xCC, []byte{'Z', '*'}}
)
var (
	SerialStatus         = RegCode{0x00, []byte{'S'}}
	SerialSpecialRelay   = RegCode{0x0100, []byte{'M'}}
	SerialInput          = RegCode{0x0080, []byte{'X'}}
	SerialOutput         = RegCode{0x00A0, []byte{'Y'}}
	SerialTimerCoil      = RegCode{0x00C0, []byte{'T', 'C'}}
	SerialCounterCoil    = RegCode{0x01C0, []byte{'C', 'C'}}
	SerialTimerValue     = RegCode{0x0800, []byte{'T', 'N'}}
	SerialCounter16Value = RegCode{0x0A00, []byte{'C', 'N'}}
	SerialCounter32Value = RegCode{0x0C00, []byte{'C', 'N'}}
	SerialData           = RegCode{0x1000, []byte{'D'}}
)

const (
	readBatch  commandType = 0x0401
	writeBatch commandType = 0x1401
)
const (
	childCommandBit  childCommandType = 0x0001
	childCommandWord childCommandType = 0x0
)

type readVar struct {
	address uint32
	code    RegCode
	length  uint16
}
type QnaProtocolStack struct {
	deputyHeader        uint16
	networkNum          byte
	plcNum              byte
	targetIONum         uint16
	targetModuleStation byte
	dataLength          uint16
	cpuTimer            uint16
	command             commandType
	childCommand        childCommandType
	data1               readVar
}

type Qna3EAsciiProtocolStack struct {
	QnaProtocolStack
}

type Qna3EBinaryProtocolStack struct {
	QnaProtocolStack
}
type ProgramPort struct {
}

type Qna interface {
	ReadVar(isBit bool, code *RegCode, startAddress int, length int) []byte
	WriteVar(isBit bool, code *RegCode, startAddress int, length int, data []byte) []byte
	Parse(input []byte) ([]byte, error)
}

func NewQna3eASCII() Qna {
	q := Qna3EAsciiProtocolStack{QnaProtocolStack{
		deputyHeader: 0x5000, networkNum: 0x00, plcNum: 0xff, targetIONum: 0x03FF, targetModuleStation: 0x00,
		dataLength: 0, cpuTimer: 0x0010, childCommand: childCommandWord,
	}}
	return &q
}
func NewQna3eBinary() Qna {
	q := Qna3EBinaryProtocolStack{QnaProtocolStack{
		deputyHeader: 0x5000, networkNum: 0x00, plcNum: 0xff, targetIONum: 0x03FF, targetModuleStation: 0x00,
		dataLength: 0, cpuTimer: 0x0010, childCommand: childCommandWord,
	}}
	return &q
}

func NewProgramPort() Qna {
	p := ProgramPort{}
	return &p
}

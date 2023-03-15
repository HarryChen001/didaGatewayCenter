package mitsubishi

import (
	"didaGatewayCenter/dataPointDriver/plc/mitsubishi/protocolStack"
	"didaGatewayCenter/domain"
	"didaGatewayCenter/net"
	"errors"
	"fmt"
	"github.com/goburrow/serial"
	"go.uber.org/zap"
	"io"
	net2 "net"
	"os"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type mitsubishi struct {
	q            protocolStack.Qna
	isConnected  bool
	iLogU        domain.ILogUsecase
	iDTU         domain.IDataTransformUsecase
	conn         domain.Software
	portConfig   *domain.DataPointPortConfig
	timeoutCount int
	lock         sync.Mutex
}

func (s *mitsubishi) Init(portConfig *domain.DataPointPortConfig, transform domain.IDataTransformUsecase) {
	s.portConfig = portConfig
	s.iDTU = transform

	go func() {
		s.connect()
	}()

	return
}

func (s *mitsubishi) connect() {

	address := fmt.Sprintf("%s:%d", s.portConfig.Param.IP, s.portConfig.Param.PortNumber)
	//	interval := 0
	switch s.portConfig.DeviceType {
	case domain.DeviceTypeMitsubishiProgramPort:
		s.q = protocolStack.NewProgramPort()
	case domain.DeviceTypeMitsubishiComputerLink:
	case domain.DeviceTypeMcBinaryQna1E:
	case domain.DeviceTypeMCBinaryQna3E:
		s.q = protocolStack.NewQna3eBinary()
	case domain.DeviceTypeMCAsciiQna3E:
		s.q = protocolStack.NewQna3eASCII()
	}
	s.iLogU.GetLogger().Info("mitsubishi plc is connecting", zap.String("portName", s.portConfig.PortName))
	if s.portConfig.PortType == domain.SerialType {
		param := s.portConfig.Param
		portNode := ""
		if runtime.GOOS == "windows" {
			portNode = fmt.Sprintf("\\\\.\\COM%d", param.COM)
		} else {
			portNode = fmt.Sprintf("/dev/COM%d", param.COM)
		}
		c := serial.Config{
			Address:  portNode,
			BaudRate: param.BandRate,
			DataBits: param.DateBits,
			StopBits: param.StopBit,
			Parity:   string(param.Parity[0]),
			Timeout:  time.Millisecond * 50,
		}
		fb, err := serial.Open(&c)
		if err != nil {
			s.iLogU.GetLogger().Warn("Failed to open serial port", zap.String("portName", s.portConfig.PortName), zap.String("node", portNode), zap.Error(err))
			return
		}
		s.isConnected = true
		s.iLogU.GetLogger().Info("mitsubishi serial port connected", zap.String("portName", s.portConfig.PortName), zap.String("node", portNode))
		ss := net.New(fb)
		s.conn = ss
	} else {
		for {
			if s.isConnected {
				time.Sleep(time.Second)
				continue
			}
			/*interval++
			if interval >= 30 {
				interval = 30
			}
			time.Sleep(time.Duration(interval) * time.Second)*/
			tcpConn, err := net.Dial("tcp", address)
			if err != nil {
				time.Sleep(time.Second)
				//	s.iLogU.GetLogger().Warn("cannot connect to mitsubishi plc", zap.String("name", s.portConfig.PortName), zap.String("address", address), zap.Error(err))
				continue
			}
			s.iLogU.GetLogger().Info("mitsubishi plc connected", zap.String("name", s.portConfig.PortName), zap.String("address", address))
			//	interval = 0
			s.conn = tcpConn
			s.isConnected = true
		}
	}
}
func (s *mitsubishi) Read(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableList *domain.DataPointVariableList) domain.IValueType {
	if !s.isConnected {
		time.Sleep(time.Second)
		return nil
	}
	dataType := variableList.DataType
	regType := variableList.Param.RegType
	regAddr := variableList.Param.RegAddr
	length := 0
	isBit := false
	var code *protocolStack.RegCode
	switch dataType {
	case domain.VarDataTypeBool:
		length = 1
		isBit = true
	case domain.VarDataTypeBit:
		length = 1
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		length = 1
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32, domain.VarDataTypeFloat:
		length = 2
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64, domain.VarDataTypeDouble:
		length = 4
	}
	if portInfo.DeviceType == domain.DeviceTypeMitsubishiProgramPort {
		switch regType {
		case domain.RegTypeMitsubishiXRegister:
			code = &protocolStack.SerialInput
		case domain.RegTypeMitsubishiYRegister:
			code = &protocolStack.SerialOutput
		case domain.RegTypeMitsubishiMRegister:
			code = &protocolStack.SerialSpecialRelay
		case domain.RegTypeMitsubishiSRegister:
			code = &protocolStack.SerialStatus
		case domain.RegTypeMitsubishiTRegister:
			code = &protocolStack.SerialTimerCoil
		case domain.RegTypeMitsubishiCRegister:
			code = &protocolStack.SerialCounterCoil
		case domain.RegTypeMitsubishiDRegister:
			code = &protocolStack.SerialData
		case domain.RegTypeMitsubishiTVRegister:
			code = &protocolStack.SerialTimerValue
		case domain.RegTypeMitsubishiCVRegister:
			code = &protocolStack.SerialCounter16Value
			if variableList.Param.RegAddr > 200 {
				code = &protocolStack.SerialCounter32Value
			}
		}
	} else {
		switch regType {
		case domain.RegTypeMitsubishiXRegister:
			code = &protocolStack.Input
		case domain.RegTypeMitsubishiYRegister:
			code = &protocolStack.Output
		case domain.RegTypeMitsubishiMRegister:
			code = &protocolStack.InternalRelay
		case domain.RegTypeMitsubishiSRegister:
			code = &protocolStack.SpecialRelay
		case domain.RegTypeMitsubishiTRegister:
			code = &protocolStack.TimerCoil
		case domain.RegTypeMitsubishiCRegister:
			code = &protocolStack.CounterCoil
		case domain.RegTypeMitsubishiDRegister:
			code = &protocolStack.DataRegister
		case domain.RegTypeMitsubishiTVRegister:
			code = &protocolStack.TimerCurrent
		case domain.RegTypeMitsubishiCVRegister:
			code = &protocolStack.CounterCurrent
		}
	}

	bb := s.q.ReadVar(isBit, code, regAddr, length)

	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.conn.WriteReadTimeout(bb, time.Second)

	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			s.timeoutCount++
			if s.timeoutCount == 5 {
				s.timeoutCount = 0
				s.isConnected = false
				s.iLogU.GetLogger().Warn("mitsubishi timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				s.iLogU.GetLogger().Warn("read from mitsubishi timeout", zap.String("portName", portInfo.PortName),
					zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
			}
			return nil
		}
		if opErr, ok := err.(*net2.OpError); ok {
			switch t := opErr.Err.(type) {
			case *os.SyscallError:
				if errno, ok := t.Err.(syscall.Errno); ok {
					switch errno {
					case syscall.ECONNABORTED:
						s.isConnected = false
						s.iLogU.GetLogger().Warn("mitsubishi connection aborted,reconnecting")
					default:
						s.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading mitsubishi", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
					}
				}
			default:
				s.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.ValueOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			s.iLogU.GetLogger().Warn("error occurred while reading mitsubishi", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
		}
	}

	valueByte, err := s.q.Parse(r)
	if err != nil {
		s.iLogU.GetLogger().Warn("parse data failed", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
		return nil
	}
	if len(valueByte) == 1 {
		valueByte = append([]byte{0}, valueByte[0])
	}
	switch code {
	case &protocolStack.SerialInput, &protocolStack.SerialOutput:
		variableList.DataType = domain.VarDataTypeBit
		variableList.Param.RegAddr = regAddr / 10
		variableList.Param.BitAddr = regAddr % 10
	case &protocolStack.SerialSpecialRelay:
		variableList.DataType = domain.VarDataTypeBit
		variableList.Param.RegAddr = regAddr / 8
		variableList.Param.BitAddr = regAddr % 8
	}
	value, err := s.iDTU.ByteToValue(deviceInfo, variableList, valueByte)
	if err != nil {
		s.iLogU.GetLogger().Warn("Error converting variable value", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
	}
	return value
}

func (s *mitsubishi) Write(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableInfo *domain.DataPointVariableList, inputValue interface{}) error {
	if !s.isConnected {
		time.Sleep(time.Second)
		return nil
	}
	dataType := variableInfo.DataType
	regType := variableInfo.Param.RegType
	regAddr := variableInfo.Param.RegAddr
	length := 0
	isBit := false
	var code *protocolStack.RegCode
	switch dataType {
	case domain.VarDataTypeBool:
		length = 1
		isBit = true
	case domain.VarDataTypeBit:
		length = 1
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		length = 1
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32, domain.VarDataTypeFloat:
		length = 2
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64, domain.VarDataTypeDouble:
		length = 4
	}
	if portInfo.DeviceType == domain.DeviceTypeMitsubishiProgramPort {
		switch regType {
		case domain.RegTypeMitsubishiXRegister:
			code = &protocolStack.SerialInput
			//	variableInfo.DataType = domain.VarDataTypeBit
			variableInfo.Param.BitAddr = regAddr % 10
		case domain.RegTypeMitsubishiYRegister:
			code = &protocolStack.SerialOutput
			//	variableInfo.DataType = domain.VarDataTypeBit
			variableInfo.Param.BitAddr = regAddr % 10
		case domain.RegTypeMitsubishiMRegister:
			code = &protocolStack.SerialSpecialRelay
			//	variableInfo.DataType = domain.VarDataTypeBit
			variableInfo.Param.BitAddr = regAddr % 10
		case domain.RegTypeMitsubishiSRegister:
			code = &protocolStack.SerialStatus
		case domain.RegTypeMitsubishiTRegister:
			code = &protocolStack.SerialTimerCoil
		case domain.RegTypeMitsubishiCRegister:
			code = &protocolStack.SerialCounterCoil
		case domain.RegTypeMitsubishiDRegister:
			code = &protocolStack.SerialData
		case domain.RegTypeMitsubishiTVRegister:
			code = &protocolStack.SerialTimerValue
		case domain.RegTypeMitsubishiCVRegister:
			code = &protocolStack.SerialCounter16Value
		}
		if variableInfo.DataType == domain.VarDataTypeBit {
			v := s.Read(portInfo, deviceInfo, variableInfo).OriginalValue()
			if inputValue.(float64) != 0 {
				inputValue = float64(v | (1 << variableInfo.Param.BitAddr))
			} else {
				inputValue = float64(v & uint16(^(0x0001 << variableInfo.Param.BitAddr)))
			}
		}
	} else {
		switch regType {
		case domain.RegTypeMitsubishiXRegister:
			code = &protocolStack.Input
		case domain.RegTypeMitsubishiYRegister:
			code = &protocolStack.Output
		case domain.RegTypeMitsubishiMRegister:
			code = &protocolStack.InternalRelay
		case domain.RegTypeMitsubishiSRegister:
			code = &protocolStack.SpecialRelay
		case domain.RegTypeMitsubishiTRegister:
			code = &protocolStack.TimerCoil
		case domain.RegTypeMitsubishiCRegister:
			code = &protocolStack.CounterCoil
		case domain.RegTypeMitsubishiDRegister:
			code = &protocolStack.DataRegister
		case domain.RegTypeMitsubishiTVRegister:
			code = &protocolStack.TimerCurrent
		case domain.RegTypeMitsubishiCVRegister:
			code = &protocolStack.CounterCurrent
		}
	}

	result, err := s.iDTU.ValueToByte(deviceInfo, variableInfo, inputValue.(float64))
	if err != nil {
		panic(err)
	}
	if dataType == domain.VarDataTypeBit {
		result = []byte{result[1]}
	}
	r1 := s.q.WriteVar(isBit, code, regAddr, length, result)

	s.lock.Lock()
	defer s.lock.Unlock()
	r, err := s.conn.WriteReadTimeout(r1, time.Second)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			s.timeoutCount++
			if s.timeoutCount == 5 {
				s.timeoutCount = 0
				s.isConnected = false
				s.iLogU.GetLogger().Warn("mitsubishi timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				s.iLogU.GetLogger().Warn("read from mitsubishi timeout", zap.String("portName", portInfo.PortName),
					zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
			}
			return nil
		} else if errors.Is(err, io.EOF) {
			s.isConnected = false
			s.iLogU.GetLogger().Warn("mitsubishi timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
		}
		if opErr, ok := err.(*net2.OpError); ok {
			switch t := opErr.Err.(type) {
			case *os.SyscallError:
				if errno, ok := t.Err.(syscall.Errno); ok {
					switch errno {
					case syscall.ECONNABORTED, io.EOF:
						s.isConnected = false
						s.iLogU.GetLogger().Warn("mitsubishi connection aborted,reconnecting")
					default:
						s.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading mitsubishi", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
					}
				}
			default:
				s.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.ValueOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			s.iLogU.GetLogger().Warn("error occurred while reading mitsubishi", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
		}
	}
	if _, err := s.q.Parse(r); err != nil {
		s.iLogU.GetLogger().Warn("mitsubishi write failed", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name),
			zap.String("result", fmt.Sprintf("% X", r)), zap.Error(err))
	}
	return nil
}

func NewMitsubishiUsecaseDriver(iLogU domain.ILogUsecase) domain.IDataPointDriverUsecase {
	m := mitsubishi{
		iLogU: iLogU,
	}
	return &m
}

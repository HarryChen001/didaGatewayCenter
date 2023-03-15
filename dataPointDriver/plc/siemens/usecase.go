package siemens

import (
	"didaGatewayCenter/dataPointDriver/plc/siemens/protocolStack"
	"didaGatewayCenter/domain"
	"didaGatewayCenter/net"
	"errors"
	"fmt"
	"go.uber.org/zap"
	net2 "net"
	"os"
	"reflect"
	"syscall"
	"time"
)

type siemens struct {
	s            *protocolStack.S7Comm
	isConnected  bool
	iLogU        domain.ILogUsecase
	iDTU         domain.IDataTransformUsecase
	conn         domain.Software
	portConfig   *domain.DataPointPortConfig
	timeoutCount int
}

func (s *siemens) Init(portConfig *domain.DataPointPortConfig, transform domain.IDataTransformUsecase) {

	s.portConfig = portConfig
	s.iDTU = transform

	go func() {
		s.connect()
	}()

	return
}
func (s *siemens) connect() {
	//TODO: set rack and slot
	s7 := protocolStack.NewS7Comm(0, 0x0a)
	s.s = s7
	b := s7.GetCoTPShakeHands(protocolStack.S200Smart)
	b2 := s7.GetCommunicationByte()

	address := fmt.Sprintf("%s:%d", s.portConfig.Param.IP, s.portConfig.Param.PortNumber)
	s.iLogU.GetLogger().Info("siemens plc is connecting", zap.String("portName", s.portConfig.PortName))
	//	interval := 0
	for {
		if s.isConnected {
			time.Sleep(time.Second)
			continue
		}
		/*	interval += 3
			if interval >= 60 {
				interval = 60
			}
			time.Sleep(time.Duration(interval) * time.Second)*/
		tcpConn, err := net.Dial("tcp", address)
		if err != nil {
			time.Sleep(time.Second)
			//	s.iLogU.GetLogger().Warn("cannot connect to plc", zap.String("name", s.portConfig.PortName), zap.String("address", address), zap.Error(err))
			continue
		}
		if _, err := tcpConn.WriteReadTimeout(b, time.Second); err != nil {
			s.iLogU.GetLogger().Warn("send cotp failed", zap.String("name", s.portConfig.PortName), zap.String("address", address), zap.Error(err))
			continue
		}
		if _, err := tcpConn.WriteReadTimeout(b2, time.Second); err != nil {
			s.iLogU.GetLogger().Warn("send setCommunication failed", zap.String("name", s.portConfig.PortName), zap.String("address", address), zap.Error(err))
			continue
		}
		s.iLogU.GetLogger().Info("siemens plc connected", zap.String("name", s.portConfig.PortName), zap.String("address", address))
		//	interval = 0
		s.conn = tcpConn
		s.isConnected = true
	}
}
func (s *siemens) Read(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableList *domain.DataPointVariableList) domain.IValueType {

	//	log.Println(variableList.Name, variableList.Id)
	if !s.isConnected {
		time.Sleep(time.Second)
		return nil
	}
	dataType := variableList.DataType
	regType := variableList.Param.RegType
	regAddr := variableList.Param.RegAddr
	dbNum := variableList.Param.DBNum
	bitAddress := variableList.Param.BitAddr

	is200family := false
	switch portInfo.DeviceType {
	case domain.DeviceTypeSiemensS200Smart, domain.DeviceTypeSiemens200CP2431:
		is200family = true
	default:
		is200family = false
	}
	sizeType, sizeCount := getTransportSize(dataType)
	area := getArea(regType, is200family)

	bb := s.s.ReadVar(sizeType, sizeCount, dbNum, area, regAddr, bitAddress)
	r, err := s.conn.WriteReadTimeout(bb, time.Second)

	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			s.timeoutCount++
			if s.timeoutCount == 5 {
				s.timeoutCount = 0
				s.isConnected = false
				s.iLogU.GetLogger().Warn("s7net timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				s.iLogU.GetLogger().Warn("read from s7net timeout", zap.String("portName", portInfo.PortName),
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
						s.iLogU.GetLogger().Warn("s7net connection aborted,reconnecting")
					default:
						s.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading s7net", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
					}
				}
			default:
				s.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.ValueOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			s.iLogU.GetLogger().Warn("error occurred while reading s7net", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
		}
	}
	valueByte, err := s.s.Parse(r)
	if err != nil {
		s.iLogU.GetLogger().Warn("parse data failed", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
		return nil
	}
	if len(valueByte) == 1 {
		valueByte = append([]byte{0}, valueByte[0])
	}
	value, err := s.iDTU.ByteToValue(deviceInfo, variableList, valueByte)
	if err != nil {
		s.iLogU.GetLogger().Warn("Error converting variable value", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
	}
	return value
}

func (s *siemens) Write(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableInfo *domain.DataPointVariableList, inputValue interface{}) error {
	if !s.isConnected {
		time.Sleep(time.Second)
		return fmt.Errorf("s7net device is not connected")
	}

	dataType := variableInfo.DataType
	regType := variableInfo.Param.RegType
	regAddr := variableInfo.Param.RegAddr
	dbNum := variableInfo.Param.DBNum
	bitAddress := variableInfo.Param.BitAddr

	is200family := false
	switch portInfo.DeviceType {
	case domain.DeviceTypeSiemensS200Smart, domain.DeviceTypeSiemens200CP2431:
		is200family = true
	default:
		is200family = false
	}
	sizeType, sizeCount := getTransportSize(dataType)
	area := getArea(regType, is200family)

	result, err := s.iDTU.ValueToByte(deviceInfo, variableInfo, inputValue.(float64))
	if err != nil {
		panic(err)
	}
	if dataType == domain.VarDataTypeBit {
		result = []byte{result[1]}
	}
	r1 := s.s.WriteVar(sizeType, sizeCount, dbNum, area, regAddr, bitAddress, result)
	r, err := s.conn.WriteReadTimeout(r1, time.Second)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			s.timeoutCount++
			if s.timeoutCount == 5 {
				s.timeoutCount = 0
				s.isConnected = false
				s.iLogU.GetLogger().Warn("s7net timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				s.iLogU.GetLogger().Warn("read from s7net timeout", zap.String("portName", portInfo.PortName),
					zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
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
						s.iLogU.GetLogger().Warn("s7net connection aborted,reconnecting")
					default:
						s.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading s7net", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
					}
				}
			default:
				s.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.ValueOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			s.iLogU.GetLogger().Warn("error occurred while reading s7net", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
		}
	}
	if _, err := s.s.Parse(r); err != nil {
		s.iLogU.GetLogger().Warn("s7net write failed", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
	}
	return nil
}

func NewSiemensDriver(iLogU domain.ILogUsecase) domain.IDataPointDriverUsecase {
	s := siemens{
		iLogU: iLogU,
	}
	return &s
}
func getTransportSize(dataType domain.DataType) (sizeType protocolStack.TransportSize, sizeCount int) {
	sizeCount = 1
	switch dataType {
	case domain.VarDataTypeBit, domain.VarDataTypeBool:
		sizeType = protocolStack.TransportSizeBit
	case domain.VarDataTypeByte:
		sizeType = protocolStack.TransportSizeByte
	case domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		sizeType = protocolStack.Word
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32, domain.VarDataTypeFloat:
		sizeType = protocolStack.TransportSizeByte
		sizeType = protocolStack.DWord
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64, domain.VarDataTypeDouble:
		sizeType = protocolStack.DWord
		sizeCount = 2
	}
	return
}
func getArea(regType domain.RegisterType, is200family bool) (area protocolStack.Area) {

	switch regType {
	case domain.RegTypeSiemensI:
		area = protocolStack.AreaTypeINPUTS
	case domain.RegTypeSiemensQ:
		area = protocolStack.AreaTypeOUTPUTS
	case domain.RegTypeSiemensV:
		area = protocolStack.AreaTypeV
	case domain.RegTypeSiemensDB:
		area = protocolStack.AreaTypeDB
	case domain.RegTypeSiemensC:
		area = protocolStack.AreaTypeCOUNTER
		if is200family {
			area = protocolStack.AreaTypeCOUNTER200
		}
	case domain.RegTypeSiemensT:
		area = protocolStack.AreaTypeTIMER
		if is200family {
			area = protocolStack.AreaTypeTIMER200
		}
	case domain.RegTypeSiemensAI:
		area = protocolStack.AreaTypeANAIN
	case domain.RegTypeSiemensAQ:
		area = protocolStack.AreaTypeANAOUT
	case domain.RegTypeSiemensM:
		area = protocolStack.AreaTypeFLAGS
	case domain.RegTypeSiemensSM:
		area = protocolStack.AreaTypeSYSFlags
	}
	return
}

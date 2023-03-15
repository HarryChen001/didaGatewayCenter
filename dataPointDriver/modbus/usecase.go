package modbus

import (
	"didaGatewayCenter/dataPointConfig/usecase"
	"didaGatewayCenter/domain"
	"errors"
	"fmt"
	"github.com/HarryChen001/go-modbus"
	"github.com/goburrow/serial"
	"go.uber.org/zap"
	"net"
	"os"
	"reflect"
	"sync"
	"syscall"
	"time"
)

type modbusDriver struct {
	iLogU            domain.ILogUsecase
	rtuClientHandler *modbus.RTUClientHandler
	tcpClientHandler *modbus.TCPClientHandler
	isConnected      bool
	lock             sync.Mutex
	modbusClient     modbus.Client
	dataTransform    domain.IDataTransformUsecase
	timeoutCount     int
}

func (m *modbusDriver) Read(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableList *domain.DataPointVariableList) domain.IValueType {

	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.isConnected {
		time.Sleep(time.Second)
		return nil
	}
	slaveDeviceAddress := deviceInfo.DevAddr
	dataType := variableList.DataType
	regType := variableList.Param.RegType
	regAddr := variableList.Param.RegAddr
	length := uint16(0)
	if portInfo.PortType == domain.NetType {
		m.tcpClientHandler.SlaveId = uint8(slaveDeviceAddress)
	} else {
		m.rtuClientHandler.SlaveId = uint8(slaveDeviceAddress)
	}
	switch dataType {
	case domain.VarDataTypeBit:
		length = 1
	case domain.VarDataTypeBool, domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		length = 1
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32, domain.VarDataTypeFloat:
		length = 2
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64, domain.VarDataTypeDouble:
		length = 4
	}
	var result []byte
	var err error
	switch regType {
	case domain.RegTypeCoilStatusWithWriteSingle, domain.RegTypeCoilStatusWithWriteMultiple:
		if dataType == domain.VarDataTypeBit {
			length = 8
			regAddr *= 8
		}
		result, err = m.modbusClient.ReadCoils(uint16(regAddr), length)
	case domain.RegTypeInputStatus:
		result, err = m.modbusClient.ReadDiscreteInputs(uint16(regAddr), length)
	case domain.RegTypeHoldingRegisterWithWriteSingle, domain.RegTypeHoldingRegisterWithWriteMultiple:
		result, err = m.modbusClient.ReadHoldingRegisters(uint16(regAddr), length)
	case domain.RegTypeInputRegister:
		result, err = m.modbusClient.ReadInputRegisters(uint16(regAddr), length)
	}
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			m.timeoutCount++
			if m.timeoutCount == 5 {
				m.timeoutCount = 0
				m.isConnected = false
				m.iLogU.GetLogger().Warn("modbus timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				m.iLogU.GetLogger().Warn("read from modbus timeout", zap.String("portName", portInfo.PortName),
					zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
			}
			return nil
		}
		if opErr, ok := err.(*net.OpError); ok {
			switch t := opErr.Err.(type) {
			case *os.SyscallError:
				if errno, ok := t.Err.(syscall.Errno); ok {
					switch errno {
					case syscall.ECONNABORTED:
						m.isConnected = false
						m.iLogU.GetLogger().Warn("modbus connection aborted,reconnecting")
					default:
						m.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading modbus", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
					}
				}
			default:
				m.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.TypeOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			m.iLogU.GetLogger().Warn("error occurred while reading modbus", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
		}
	}
	if len(result) == 1 {
		result = []byte{0, result[0]}
	}
	value, err := m.dataTransform.ByteToValue(deviceInfo, variableList, result)
	if err != nil {
		m.iLogU.GetLogger().Warn("Error converting variable value", zap.String("portName", portInfo.PortName),
			zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableList.Name), zap.Error(err))
	}

	return value
}

func (m *modbusDriver) Write(portInfo *domain.DataPointPortConfig, deviceInfo *domain.DeviceList, variableInfo *domain.DataPointVariableList, inputValue interface{}) error {
	if !m.isConnected {
		time.Sleep(time.Second)
		return nil
	}
	slaveDeviceAddress := deviceInfo.DevAddr
	dataType := variableInfo.DataType
	regType := variableInfo.Param.RegType
	regAddr := variableInfo.Param.RegAddr
	length := uint16(0)
	if portInfo.PortType == domain.NetType {
		m.tcpClientHandler.SlaveId = uint8(slaveDeviceAddress)
	} else {
		m.rtuClientHandler.SlaveId = uint8(slaveDeviceAddress)
	}
	value := (inputValue.(float64) - variableInfo.Offset) / variableInfo.Modulus
	switch dataType {
	case domain.VarDataTypeBit:
		v := m.Read(portInfo, deviceInfo, variableInfo).OriginalValue()
		if inputValue.(float64) != 0 {
			value = float64(v | (1 << variableInfo.Param.BitAddr))
		} else {
			value = float64(v & uint16(^(0x0001 << variableInfo.Param.BitAddr)))
		}
		length = 1
	case domain.VarDataTypeBool, domain.VarDataTypeUint16, domain.VarDataTypeInt16:
		length = 1
	case domain.VarDataTypeUint32, domain.VarDataTypeInt32, domain.VarDataTypeFloat:
		length = 2
	case domain.VarDataTypeUint64, domain.VarDataTypeInt64, domain.VarDataTypeDouble:
		length = 4
	}
	result, err := m.dataTransform.ValueToByte(deviceInfo, variableInfo, value)

	switch regType {
	case domain.RegTypeCoilStatusWithWriteSingle:
		temp := uint16(0)
		if inputValue.(float64) != 0 {
			temp = 0xff00
		}
		if dataType == domain.VarDataTypeBit {
			regAddr = regAddr*8 + variableInfo.Param.BitAddr
		}
		result, err = m.modbusClient.WriteSingleCoil(uint16(regAddr), temp)
	case domain.RegTypeCoilStatusWithWriteMultiple:
		if dataType == domain.VarDataTypeBit {
			length = 8
			regAddr *= 8
		}
		result, err = m.modbusClient.WriteMultipleCoils(uint16(regAddr), length, []byte{result[1]})
	case domain.RegTypeHoldingRegisterWithWriteSingle:
		result, err = m.modbusClient.WriteSingleRegister(uint16(regAddr), uint16(int16(value)))
	case domain.RegTypeHoldingRegisterWithWriteMultiple:
		result, err = m.modbusClient.WriteMultipleRegisters(uint16(regAddr), length, result)
	}
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			m.timeoutCount++
			if m.timeoutCount == 5 {
				m.timeoutCount = 0
				m.isConnected = false
				m.iLogU.GetLogger().Warn("modbus timeout times have exceeded the limit, reconnecting", zap.String("portName", portInfo.PortName))
			} else {
				m.iLogU.GetLogger().Warn("write modbus timeout", zap.String("portName", portInfo.PortName),
					zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
			}
			return nil
		}
		if opErr, ok := err.(*net.OpError); ok {
			switch t := opErr.Err.(type) {
			case *os.SyscallError:
				if errno, ok := t.Err.(syscall.Errno); ok {
					switch errno {
					case syscall.ECONNABORTED:
						m.isConnected = false
						m.iLogU.GetLogger().Warn("modbus connection aborted,reconnecting")
					default:
						m.iLogU.GetLogger().Warn("unexpected syscall error occurred while reading modbus", zap.String("portName", portInfo.PortName),
							zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
					}
				}
			default:
				m.iLogU.GetLogger().Warn("unknown operation error", zap.String("error type", reflect.TypeOf(opErr.Err).String()), zap.Error(err))
			}
		} else {
			m.iLogU.GetLogger().Warn("error occurred while reading modbus", zap.String("portName", portInfo.PortName),
				zap.String("deviceName", deviceInfo.DevName), zap.String("variableName", variableInfo.Name), zap.Error(err))
		}
	}
	return nil
}

func NewModbusUsecase(iLU domain.ILogUsecase) domain.IDataPointDriverUsecase {
	m := &modbusDriver{iLogU: iLU}
	return m
}

func (m *modbusDriver) Init(portConfig *domain.DataPointPortConfig, transform domain.IDataTransformUsecase) {
	portName := portConfig.PortName
	m.dataTransform = transform
	if portConfig.PortType == domain.SerialType {
		comNum := portConfig.Param.COM
		deviceNode := usecase.ConvertComToDeviceNode(comNum)
		rtuClientHandler := modbus.NewRTUClientHandler(deviceNode)
		rtuClientHandler.Config = serial.Config{
			Address:  deviceNode,
			BaudRate: portConfig.Param.BandRate,
			DataBits: portConfig.Param.DateBits,
			StopBits: portConfig.Param.StopBit,
			Parity:   string(portConfig.Param.Parity[0]),
			Timeout:  time.Duration(portConfig.Param.RespTimeOutMs) * time.Millisecond,
		}
		if err := rtuClientHandler.Connect(); err != nil {
			m.iLogU.GetLogger().Error("modbus rtu connect failed", zap.String("port", portName), zap.String("deviceNode", deviceNode), zap.Error(err))
		} else {
			m.isConnected = true
			m.iLogU.GetLogger().Info("modbus rtu connect succeeded", zap.String("port", portName), zap.String("deviceNode", deviceNode))
		}
		m.rtuClientHandler = rtuClientHandler
		client := modbus.NewClient(rtuClientHandler)
		m.modbusClient = client
	} else {
		deviceNode := fmt.Sprintf("%s:%d", portConfig.Param.IP, portConfig.Param.PortNumber)
		tcpClientHandler := modbus.NewTCPClientHandler(deviceNode)
		if err := tcpClientHandler.Connect(); err != nil {
			m.iLogU.GetLogger().Error("modbus tcp connect failed", zap.String("port", portName), zap.String("deviceNode", deviceNode), zap.Error(err))
		} else {
			m.isConnected = true
			m.iLogU.GetLogger().Info("modbus tcp connect succeeded", zap.String("port", portName), zap.String("deviceNode", deviceNode))
		}
		m.tcpClientHandler = tcpClientHandler
		client := modbus.NewClient(tcpClientHandler)
		m.modbusClient = client
		go func() {
			for {
				if m.isConnected {
					time.Sleep(time.Second * 2)
					continue
				}
				if m.tcpClientHandler != nil {
					tcpClientHandler = modbus.NewTCPClientHandler(deviceNode)
					if err := tcpClientHandler.Connect(); err != nil {
						time.Sleep(time.Second * 2)
						continue
					} else {
						m.isConnected = true
						m.tcpClientHandler = tcpClientHandler
						c := modbus.NewClient(tcpClientHandler)
						m.modbusClient = c
						m.iLogU.GetLogger().Info("modbus tcp reconnect success", zap.String("port", portName), zap.String("deviceNode", deviceNode))
					}
				}
			}
		}()
	}
}

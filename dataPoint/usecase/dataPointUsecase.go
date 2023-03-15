package usecase

import (
	"didaGatewayCenter/dataPointDriver/modbus"
	"didaGatewayCenter/dataPointDriver/plc/mitsubishi"
	"didaGatewayCenter/dataPointDriver/plc/siemens"
	"didaGatewayCenter/dataTransform"
	"didaGatewayCenter/domain"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type dataPointUsecase struct {
	dataPoints []*domain.DataPoint
	logUsecase domain.ILogUsecase
}

func (d *dataPointUsecase) WriteById(id int64, value interface{}) (interface{}, error) {
	portConfig := &domain.DataPointPortConfig{}
	deviceList := &domain.DeviceList{}
	var variableList *domain.DataPointVariableList
	var driver domain.IDataPointDriverUsecase
	//dataPointIndex := 0
	portConfig, deviceList, variableList, driver = d.findVariableById(id)

	if variableList == nil {
		return nil, fmt.Errorf("variable name is not found")
	}
	if err := driver.Write(portConfig, deviceList, variableList, value); err != nil {
		//	if err := d.dataPoints[dataPointIndex].Driver.Write(portConfig, deviceList, variableList, value); err != nil {
		return nil, err
	}
	value, err := d.ReadById(id, true)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (d *dataPointUsecase) GetStore() []domain.AllDataPoints {
	var ret []domain.AllDataPoints
	for _, singleDataPoint := range d.dataPoints {
		for _, singleVariable := range singleDataPoint.VariableConfig {
			for _, singleVariableList := range singleVariable.VarList {
				ret = append(ret, domain.AllDataPoints{
					PortName:     singleVariable.PortName,
					DeviceName:   singleVariable.DevName,
					VariableName: singleVariableList.Name,
					Value:        singleVariableList.Value,
					Timestamp:    singleVariableList.Timestamp,
				})
			}
		}
	}
	return ret
}

func NewDataPointUseCase(logUc domain.ILogUsecase, dataPointConfig domain.IDataPointConfigUseCase) domain.IDataPointUseCase {
	d := &dataPointUsecase{
		logUsecase: logUc,
	}
	dataPointPorts := dataPointConfig.GetPortConfigs()

	var dataPointDriver domain.IDataPointDriverUsecase

	for index, singleDataPointPort := range dataPointPorts.PortConfigs {
		if !singleDataPointPort.Vaild {
			logUc.GetLogger().Info("the port is disabled,skipping", zap.String("port", singleDataPointPort.PortName))
			continue
		}
		portName := singleDataPointPort.PortName
		tempDataPoint := domain.DataPoint{}
		tempDataPoint.PortConfig = &dataPointPorts.PortConfigs[index]
		tempDataPoint.DeviceConfig = dataPointConfig.GetDeviceConfigs(portName)
		tempDataPoint.VariableConfig = dataPointConfig.GetVariableConfigs(portName)
		switch singleDataPointPort.DeviceType {
		case domain.DeviceTypeModbusRTU, domain.DeviceTypeModbusTCP, domain.DeviceTypeModbusASCII:
			dataPointDriver = modbus.NewModbusUsecase(logUc)
		case domain.DeviceTypeSiemensS200Smart:
			dataPointDriver = siemens.NewSiemensDriver(logUc)
		case domain.DeviceTypeMCAsciiQna3E, domain.DeviceTypeMCBinaryQna3E, domain.DeviceTypeMitsubishiProgramPort,
			domain.DeviceTypeMitsubishiComputerLink:
			dataPointDriver = mitsubishi.NewMitsubishiUsecaseDriver(logUc)
		default:
			continue
		}
		tempDataPoint.Driver = dataPointDriver
		d.dataPoints = append(d.dataPoints, &tempDataPoint)
		dataPointDriver.Init(&dataPointPorts.PortConfigs[index], dataTransform.NewDataTransformUsecase())
	}
	return d
}
func (d *dataPointUsecase) Read(portName string, deviceName string, variableName string, isRealTime bool) (interface{}, error) {
	portConfig := &domain.DataPointPortConfig{}
	deviceList := &domain.DeviceList{}
	var variableList *domain.DataPointVariableList
	dataPointIndex := 0
	for index, singleDataPoint := range d.dataPoints {
		if portName != "" && singleDataPoint.PortConfig.PortName != portName {
			continue
		}
		portConfig = singleDataPoint.PortConfig
		for _, singleDeviceList := range singleDataPoint.DeviceConfig.DevList {
			if deviceName != "" && singleDeviceList.DevName != deviceName {
				continue
			}
			deviceList = singleDeviceList
		}

		for _, singleVariableConfig := range singleDataPoint.VariableConfig {

			if singleVariableConfig.PortName != portName {
				continue
			} else if singleVariableConfig.DevName != deviceName {
				continue
			}
			for _, singleVariableList := range singleVariableConfig.VarList {
				if singleVariableList.Name != variableName {
					continue
				}
				if variableList == nil {
					variableList = &domain.DataPointVariableList{}
				}
				variableList = &singleVariableList
				dataPointIndex = index
			}
			break
		}
	}
	if variableList == nil {
		return nil, fmt.Errorf("variable name is not found")
	}
	var value interface{}
	if isRealTime {

		value = d.dataPoints[dataPointIndex].Driver.Read(portConfig, deviceList, variableList).ToFloat64()
	} else {
		value = variableList.Value
	}
	return value, nil
}

func (d *dataPointUsecase) ReadById(id int64, isRealTime bool) (interface{}, error) {
	portConfig := &domain.DataPointPortConfig{}
	deviceList := &domain.DeviceList{}

	var (
		variableList *domain.DataPointVariableList
		driver       domain.IDataPointDriverUsecase
	)

	//	dataPointIndex := 0
	portConfig, deviceList, variableList, driver = d.findVariableById(id)
	if variableList == nil {
		return nil, fmt.Errorf("variable name is not found")
	}
	var value interface{}
	if isRealTime {
		tempValue := driver.Read(portConfig, deviceList, variableList)
		if tempValue == nil {
			return nil, nil
		}
		return tempValue.ToFloat64(), nil
	} else {
		value = variableList.Value
	}
	return value, nil
}

func (d *dataPointUsecase) findVariableById(id int64) (*domain.DataPointPortConfig, *domain.DeviceList, *domain.DataPointVariableList, domain.IDataPointDriverUsecase) {
	portName := ""
	deviceName := ""
	portConfig := &domain.DataPointPortConfig{}
	devList := &domain.DeviceList{}
	var variableList *domain.DataPointVariableList
	var driver domain.IDataPointDriverUsecase
	for _, singleDataPoint := range d.dataPoints {
		for _, singleVariableConfig := range singleDataPoint.VariableConfig {
			for _, singleVarList := range singleVariableConfig.VarList {
				if singleVarList.Id != id {
					continue
				}
				portName = singleVariableConfig.PortName
				deviceName = singleVariableConfig.DevName
				variableList = &singleVarList
				break
			}
			if portName != "" && deviceName != "" {
				break
			}
		}
		for _, singleDevList := range singleDataPoint.DeviceConfig.DevList {
			if singleDevList.DevName != deviceName {
				continue
			}
			devList = singleDevList
			break
		}
		if portName != "" && devList.DevName != "" {
			portConfig = singleDataPoint.PortConfig
			driver = singleDataPoint.Driver
			break
		}
	}
	return portConfig, devList, variableList, driver
}
func (d *dataPointUsecase) CycleSample() {

	for index, _ := range d.dataPoints {
		go func(tempIndex int, tempSingleDataPointPort *domain.DataPoint) {
			for {
				for _, singleDeviceList := range tempSingleDataPointPort.DeviceConfig.DevList {

					for index2, singleVariableConfig := range tempSingleDataPointPort.VariableConfig {
						if singleVariableConfig.PortName != tempSingleDataPointPort.PortConfig.PortName ||
							singleVariableConfig.DevName != singleDeviceList.DevName {
							continue
						}
						for index3, singleVariableList := range singleVariableConfig.VarList {

							value := tempSingleDataPointPort.Driver.Read(tempSingleDataPointPort.PortConfig, singleDeviceList, &singleVariableList)
							if value == nil {
								tempSingleDataPointPort.VariableConfig[index2].VarList[index3].Value = nil
							} else {
								tempSingleDataPointPort.VariableConfig[index2].VarList[index3].Value = value.ToFloat64()
							}
							tempSingleDataPointPort.VariableConfig[index2].VarList[index3].Timestamp = time.Now()
						}
					}
				}
				time.Sleep(time.Second * time.Duration(tempSingleDataPointPort.PortConfig.Param.SampleIntervalS))
			}
		}(index, d.dataPoints[index])
	}
}

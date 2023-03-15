package usecase

import (
	"didaGatewayCenter/domain"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

type dataPointConfigUsecase struct {
	iAppConfig domain.IAppConfigUseCase
	iLu        domain.ILogUsecase
	port       *domain.Port
	device     *domain.Device
	variable   *domain.Variable
}

func (d *dataPointConfigUsecase) GetPortConfigs() *domain.Port {
	return d.port
}

func (d *dataPointConfigUsecase) GetPortConfigsByDeviceType(deviceType domain.DeviceType) *domain.Port {

	result := domain.Port{}
	for _, singlePort := range d.port.PortConfigs {
		if singlePort.DeviceType != deviceType {
			continue
		}
		result.PortConfigs = append(result.PortConfigs, singlePort)
	}
	return &result
}

func (d *dataPointConfigUsecase) GetDeviceConfigs(portName string) *domain.DataPointDeviceConfig {

	for _, singleDevice := range d.device.DeviceConfigs {
		if singleDevice.PortName != portName {
			continue
		}
		return &singleDevice
	}
	return nil
}

func (d *dataPointConfigUsecase) GetVariableConfigs(portName string) []*domain.DataPointVariableConfig {
	var temp []*domain.DataPointVariableConfig
	for _, singleVariable := range d.variable.VariableConfigs {
		if singleVariable.PortName != portName {
			continue
		}
		temp = append(temp, &domain.DataPointVariableConfig{
			PortName: singleVariable.PortName,
			DevName:  singleVariable.DevName,
			VarList:  singleVariable.VarList,
		})
	}
	return temp
}

func ConvertComToDeviceNode(comNum int) string {
	if runtime.GOOS != "linux" {
		return fmt.Sprintf("COM%d", comNum)
	}
	return fmt.Sprintf("/dev/COM%d", comNum)
}

func NewDataPointConfigUseCase(iLogU domain.ILogUsecase, iAu domain.IAppConfigUseCase) domain.IDataPointConfigUseCase {

	dataPointPort := domain.Port{}
	dataPointDevice := domain.Device{}
	dataPointVariable := domain.Variable{}

	dataPointConfig := iAu.GetAppDataPointConfig()
	dataPointConfigPath := dataPointConfig.Path

	portFileLocation := path.Join(dataPointConfigPath, "PORTConfig.json")

	portInfo, err := os.ReadFile(portFileLocation)
	if err != nil {
		iLogU.GetLogger().Warn("no port config file found", zap.String("file", portFileLocation))
	} else {
		if err := json.Unmarshal(portInfo, &dataPointPort); err != nil {
			// TODO:handle error
			panic(err)
		}
	}

	deviceFileLocation := path.Join(dataPointConfigPath, "DEVConfig.json")
	devInfo, err := os.ReadFile(deviceFileLocation)
	if err != nil {
		iLogU.GetLogger().Warn("no device config file found", zap.String("file", deviceFileLocation))
	} else {
		if err := json.Unmarshal(devInfo, &dataPointDevice); err != nil {
			// TODO:
			panic(err)
		}
	}
	dirInfo, err := os.ReadDir(dataPointConfigPath)
	if err != nil {
		iLogU.GetLogger().Warn("")
	}
	i := int64(1)
	for _, singleVariableFileInfo := range dirInfo {
		tempDataPointVariable := domain.Variable{}
		if singleVariableFileInfo.IsDir() {
			continue
		}
		fileName := singleVariableFileInfo.Name()
		if !strings.HasPrefix(fileName, "VARConfig") {
			continue
		}
		fileInfo, err := os.ReadFile(path.Join(dataPointConfigPath, fileName))
		if err != nil {
			iLogU.GetLogger().Warn("read variable config file failed", zap.String("fileName", fileName), zap.Error(err))
		}
		if err := json.Unmarshal(fileInfo, &tempDataPointVariable); err != nil {
			// TODO: error
			panic(err)
		}
		for index1, singleVariableInfo := range tempDataPointVariable.VariableConfigs {
			for index2, _ := range singleVariableInfo.VarList {
				log.Println(singleVariableInfo.VarList[index2].Name, i)
				tempDataPointVariable.VariableConfigs[index1].VarList[index2].Id = i
				i++
			}
		}
		if len(dataPointVariable.VariableConfigs) == 0 {
			dataPointVariable.VariableConfigs = append(dataPointVariable.VariableConfigs, tempDataPointVariable.VariableConfigs...)
			continue
		}
		for index, singleVariableInfo := range dataPointVariable.VariableConfigs {
			for index2, singleTempDataPointVariableInfo := range tempDataPointVariable.VariableConfigs {
				if singleVariableInfo.PortName == singleTempDataPointVariableInfo.PortName &&
					singleVariableInfo.DevName == singleTempDataPointVariableInfo.DevName {
					dataPointVariable.VariableConfigs[index].VarList = append(dataPointVariable.VariableConfigs[index].VarList, singleTempDataPointVariableInfo.VarList...)
					continue
				} else if index == len(dataPointVariable.VariableConfigs)-1 &&
					index2 == len(tempDataPointVariable.VariableConfigs)-1 {
					dataPointVariable.VariableConfigs = append(dataPointVariable.VariableConfigs, tempDataPointVariable.VariableConfigs...)
				}
			}
		}
	}
	return &dataPointConfigUsecase{
		port:     &dataPointPort,
		device:   &dataPointDevice,
		variable: &dataPointVariable,
		iLu:      iLogU,
	}
}

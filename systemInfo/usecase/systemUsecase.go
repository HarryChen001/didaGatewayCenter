package usecase

import (
	"didaGatewayCenter/domain"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"os"
)

type systemUseCase struct {
	s *domain.System
}

func (s *systemUseCase) GetMachineInfoServer() string {
	return s.s.MachineInfo.Server
}

func (s *systemUseCase) GetMachineInfoPort() int {
	return s.s.MachineInfo.Port
}

func (s *systemUseCase) GetMachineInfoEnabled() bool {
	return s.s.MachineInfo.Enabled
}

func (s *systemUseCase) GetMachineInfoUploadIntervalS() int {
	return s.s.MachineInfo.UploadIntervalS
}

func (s *systemUseCase) GetMachineInfoName() string {
	return s.s.MachineInfo.Name
}

func (s *systemUseCase) GetMachineInfoSn() string {
	return s.s.MachineInfo.Sn
}

func (s *systemUseCase) GetMachineInfoMac() string {
	return s.s.MachineInfo.Mac
}

func NewSystemUseCase(iLog domain.ILogUsecase, iACU domain.IAppConfigUseCase) domain.ISystemUseCase {
	s := &systemUseCase{
		s: &domain.System{},
	}
	logger := iLog.GetLogger()

	fileName := iACU.GetAppSystemInfo().File
	fileInfo, err := os.ReadFile(fileName)
	if err != nil {
		logger.Panic("An error Occurred while reading system info", zap.String("fileName", fileName), zap.Error(err))
	}
	if err := yaml.Unmarshal(fileInfo, &s.s.MachineInfo); err != nil {
		logger.Panic("An error occurred while unmarshalling system info", zap.String("fileName", fileName), zap.Error(err))
	}
	machineInfo := s.s.MachineInfo
	logger.Info("System information", zap.String("name", machineInfo.Name),
		zap.String("sn", machineInfo.Sn),
		zap.String("mac", machineInfo.Mac),
		zap.String("server", machineInfo.Server),
		zap.Int("port", machineInfo.Port),
		zap.Bool("enabled", machineInfo.Enabled),
		zap.Int("uploadIntervalS", machineInfo.UploadIntervalS))
	return s
}

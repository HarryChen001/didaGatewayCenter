package main

import (
	"didaGatewayCenter/api"
	"didaGatewayCenter/appConfig/usecase"
	"didaGatewayCenter/dataPoint/delivery/http"
	usecase4 "didaGatewayCenter/dataPoint/usecase"
	http2 "didaGatewayCenter/dataPointConfig/delivery/http"
	usecase3 "didaGatewayCenter/dataPointConfig/usecase"
	usecase2 "didaGatewayCenter/log/usecase"
	usecase7 "didaGatewayCenter/mqtt/usecase"
	usecase5 "didaGatewayCenter/systemInfo/usecase"
	"flag"
)

var (
	configFile = flag.String("c", "/etc/didaGateway/config.json", "specify the configuration file,default is /etc/didaGateway/config.json")
)

func main() {

	flag.Parse()

	iACU := usecase.NewAppConfigUseCase(*configFile)
	iACU.ParseConfig()
	iLogU := usecase2.NewLogUserCase(iACU)
	iLogU.GetLogger().Info("test info")

	iSU := usecase5.NewSystemUseCase(iLogU, iACU)

	iDPCU := usecase3.NewDataPointConfigUseCase(iLogU, iACU)
	iDPU := usecase4.NewDataPointUseCase(iLogU, iDPCU)
	go iDPU.CycleSample()

	iDPH := http.NewDataPointHandler(iDPU)
	iDPCH := http2.NewDataPointConfigHandler(iLogU, iACU, iDPCU)
	api.NewApiUsecase(iLogU, iACU, iDPH, iDPCH)
	//	usecase6.NewMqttUseCase(iACU, iLogU, iSU, iDPU)
	usecase7.NewMqttUseCase(iACU, iLogU, iSU, iDPU)
	select {}
}

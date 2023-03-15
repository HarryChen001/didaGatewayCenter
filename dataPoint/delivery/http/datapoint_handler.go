package http

import (
	"didaGatewayCenter/domain"
	"github.com/labstack/echo"
)

type DataPointHandler struct {
	iDPU domain.IDataPointUseCase
}

func NewDataPointHandler(useCase domain.IDataPointUseCase) domain.IDataPointHandler {
	handler := &DataPointHandler{
		iDPU: useCase,
	}
	return handler
}

func (i *DataPointHandler) GetAllVariablesV2(ctx echo.Context) error {
	a := i.iDPU.GetStore()
	ret := make(map[string]interface{})
	ret["ret"] = a
	ctx.Response().Header().Set("Content-Type", "application/json")
	return ctx.JSON(200, ret)
}

func (i *DataPointHandler) GetAllVariablesV1(ctx echo.Context) error {
	a := i.iDPU.GetStore()
	ret := make(map[string]map[string]map[string]interface{})

	for _, singleDataPoint := range a {
		if ret[singleDataPoint.PortName] == nil {
			ret[singleDataPoint.PortName] = make(map[string]map[string]interface{})
		}
		if ret[singleDataPoint.PortName][singleDataPoint.DeviceName] == nil {
			ret[singleDataPoint.PortName][singleDataPoint.DeviceName] = make(map[string]interface{})
		}
		ret[singleDataPoint.PortName][singleDataPoint.DeviceName][singleDataPoint.VariableName] = singleDataPoint.Value
	}
	ctx.Response().Header().Set("Content-Type", "application/json")
	return ctx.JSON(200, ret)
}

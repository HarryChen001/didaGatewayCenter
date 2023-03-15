package api

import (
	"didaGatewayCenter/domain"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

type Api struct {
	iLU   domain.ILogUsecase
	iACU  domain.IAppConfigUseCase
	iDPH  domain.IDataPointHandler
	iDPCH domain.IDataPointConfigHandler
	e     *echo.Echo
}

func (a *Api) GET(path string, handlerFunc echo.HandlerFunc, middlewareFunc ...echo.MiddlewareFunc) *echo.Route {
	return a.e.GET(path, handlerFunc, middlewareFunc...)
}

func (a *Api) POST(path string, handlerFunc echo.HandlerFunc, middlewareFunc ...echo.MiddlewareFunc) *echo.Route {
	return a.e.POST(path, handlerFunc, middlewareFunc...)
}

func NewApiUsecase(logUsecase domain.ILogUsecase, appConfigUsecase domain.IAppConfigUseCase, dataPointHandler domain.IDataPointHandler, dataPointConfigHandler domain.IDataPointConfigHandler) domain.IApiUsecase {
	d := &Api{
		iLU:   logUsecase,
		iACU:  appConfigUsecase,
		iDPH:  dataPointHandler,
		iDPCH: dataPointConfigHandler,
	}
	a := d.iACU.GetConfig().Server.Address
	e := echo.New()
	{
		e.GET("/v1/checkConnect", func(context echo.Context) error {
			context.JSON(http.StatusOK, nil)
			return nil
		})
		e.GET("/v1/getAllVariables", dataPointHandler.GetAllVariablesV1)
		e.POST("/v1/configUpdate", dataPointConfigHandler.ConfigUpdate)
	}
	{
		e.GET("/v2/getAllVariables", dataPointHandler.GetAllVariablesV2)
	}
	d.e = e
	e.HideBanner = true
	go func() {
		err := e.Start(a)
		if err != nil {
			log.Println(err)
		}
	}()
	return d
}

package domain

import "github.com/labstack/echo"

type Api struct {
	Code  int         `json:"code"`
	Msg   interface{} `json:"msg"`
	Error interface{} `json:"error"`
}

type IApiUsecase interface {
	GET(path string, handlerFunc echo.HandlerFunc, middlewareFunc ...echo.MiddlewareFunc) *echo.Route
	POST(path string, handlerFunc echo.HandlerFunc, middlewareFunc ...echo.MiddlewareFunc) *echo.Route
}

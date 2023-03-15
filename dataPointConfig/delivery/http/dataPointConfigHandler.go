package http

import (
	"didaGatewayCenter/domain"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type dataPointConfigHandler struct {
	iLogU domain.ILogUsecase
	iDPC  domain.IDataPointConfigUseCase
	iACU  domain.IAppConfigUseCase
}

func NewDataPointConfigHandler(iLogU domain.ILogUsecase, iACU domain.IAppConfigUseCase, useCase domain.IDataPointConfigUseCase) domain.IDataPointConfigHandler {
	handler := &dataPointConfigHandler{
		iLogU: iLogU,
		iACU:  iACU,
		iDPC:  useCase,
	}
	return handler
}

func (d *dataPointConfigHandler) ConfigUpdate(ctx echo.Context) error {

	ret := domain.Api{
		Code:  0,
		Msg:   nil,
		Error: nil,
	}
	var err error
	httpStatus := http.StatusOK
	defer func() {
		ctx.JSON(httpStatus, ret)
	}()
	dataPointConfigPath := d.iACU.GetAppDataPointConfig().Path
	form, err := ctx.MultipartForm()
	if err != nil {
		ret.Code = -1
		ret.Msg = "获取文件内容失败"
		d.iLogU.GetLogger().Error("get file from multipart failed", zap.Error(err))
		httpStatus = http.StatusBadRequest
		return err
	}
	d.iLogU.GetLogger().Debug("get file from multipart success")
	tempDir := path.Join("./temp/tempConfigDownload/", "tempConfig")
	if err = os.RemoveAll(tempDir); err != nil {
		ret.Code = -1
		ret.Msg = "删除临时目录失败"
		d.iLogU.GetLogger().Error("delete temporary config directory failed", zap.Error(err), zap.String("path", tempDir))

		httpStatus = http.StatusInternalServerError
		return err
	}
	d.iLogU.GetLogger().Debug("delete temporary config directory success", zap.String("path", tempDir))
	if err = os.MkdirAll(tempDir, 0644); err != nil {
		ret.Code = -1
		ret.Msg = "创建临时目录失败"
		d.iLogU.GetLogger().Error("create temporary config directory failed", zap.Error(err), zap.String("path", tempDir))
		httpStatus = http.StatusInternalServerError
		return err
	}
	d.iLogU.GetLogger().Debug("create temporary config directory success", zap.String("path", tempDir))

	count := 0

	buffers := make([]byte, 100)
	for key, fileHeaders := range form.File {
		finalTempDir := tempDir
		switch key {
		case "publish":
			finalTempDir = path.Join(tempDir, "MqttMessage", "publish")
		case "subscribe":
			finalTempDir = path.Join(tempDir, "MqttMessage", "subscribe")
		}
		if err = os.MkdirAll(finalTempDir, 0755); err != nil {
			d.iLogU.GetLogger().Error("create directory for storing msg format files failed", zap.String("dirname", finalTempDir), zap.Error(err))
		} else {
			d.iLogU.GetLogger().Debug("create directory for storing msg format files success", zap.String("dirname", finalTempDir))
		}

		for _, singleFile := range fileHeaders {
			count++
			dstFd, err := os.Create(path.Join(finalTempDir, singleFile.Filename))
			if err != nil {
				ret.Code = -1
				ret.Msg = "创建临时文件失败"
				d.iLogU.GetLogger().Error("create temporary file failed", zap.Error(err), zap.String("filename", singleFile.Filename))
				httpStatus = http.StatusInternalServerError
				return err
			}
			d.iLogU.GetLogger().Debug("created temporary file", zap.String("filename", singleFile.Filename))
			srcFd, err := singleFile.Open()
			if err != nil {
				ret.Code = -1
				ret.Msg = "打开上传文件失败"
				d.iLogU.GetLogger().Error("open upload file failed", zap.Error(err), zap.String("filename", singleFile.Filename))
				httpStatus = http.StatusInternalServerError
				return err
			}
			d.iLogU.GetLogger().Debug("open upload file success", zap.String("filename", singleFile.Filename))
		reread:
			n, err := srcFd.Read(buffers)
			if err != nil && err != io.EOF {
				ret.Code = -1
				ret.Msg = "读取上传文件失败"
				d.iLogU.GetLogger().Error("read upload file failed", zap.Error(err), zap.String("filename", singleFile.Filename))
				httpStatus = http.StatusInternalServerError
				return err
			} else if err == nil {
				if _, err = dstFd.Write(buffers[:n]); err != nil {
					ret.Code = -1
					ret.Msg = "写入到临时文件失败"
					d.iLogU.GetLogger().Error("write to temporary failed", zap.Error(err), zap.String("filename", singleFile.Filename))
					httpStatus = http.StatusInternalServerError
					return err
				}
				goto reread
			}
			if err = dstFd.Close(); err != nil {
				ret.Code = -1
				ret.Msg = "临时文件无法正常关闭，下载可能失败"
				d.iLogU.GetLogger().Error("close temporary file failed", zap.Error(err), zap.String("filename", singleFile.Filename))
				httpStatus = http.StatusInternalServerError
				return err
			}
			if err = srcFd.Close(); err != nil {
				d.iLogU.GetLogger().Warn("upload file cannot close", zap.Error(err), zap.String("filename", singleFile.Filename))
			}
			d.iLogU.GetLogger().Info("save upload file to temporary directory success", zap.String("filename", singleFile.Filename))
		}
	}
	log.Println(os.RemoveAll(dataPointConfigPath))
	if err := os.Rename(tempDir, dataPointConfigPath); err != nil {
		log.Println(err)
	}
	return nil
}

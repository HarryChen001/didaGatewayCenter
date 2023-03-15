package domain

import "go.uber.org/zap"

type ILogUsecase interface {
	GetLogger() *zap.Logger
}

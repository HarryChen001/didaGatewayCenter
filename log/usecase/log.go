package usecase

import (
	"didaGatewayCenter/domain"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path"
	"sync"
)

type logUseCase struct {
	once           *sync.Once
	iConfigUseCase domain.IAppConfigUseCase
	logger         *zap.Logger
}

func NewLogUserCase(c domain.IAppConfigUseCase) domain.ILogUsecase {
	l := &logUseCase{
		once:           new(sync.Once),
		iConfigUseCase: c,
		logger:         nil,
	}
	l.logInit()
	return l
}
func (l *logUseCase) logInit() {

	l.once.Do(
		func() {
			logConfig := l.iConfigUseCase.GetLogConfig()
			level, err := zapcore.ParseLevel(logConfig.Level)
			if err != nil {
				level = zapcore.InfoLevel
			}
			logPath := logConfig.Path
			encoderConfig := zapcore.EncoderConfig{
				MessageKey:     "msg",
				LevelKey:       "level",
				TimeKey:        "time",
				NameKey:        "logger",
				CallerKey:      "line",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.FullCallerEncoder,
				EncodeName:     zapcore.FullNameEncoder,
			}

			hostname, _ := os.Hostname()
			_ = hostname
			filename := path.Join(logPath, "didaGatewayCenter.log")
			writer := newWriter1(filename)

			encoder := zapcore.NewConsoleEncoder(encoderConfig)
			//			writerSyncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout),)

			writerSyncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout))
			/*	if runtime.GOOS == "windows" {
					writerSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout))
				} else {
					path := fmt.Sprintf("/var/log/didaGateway/")
					_ = os.MkdirAll(path, 0766)
					writer := newWriter(path + "/" + "didaGateway_main.log")
					writerSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout))
				}*/

			core := zapcore.NewCore(encoder, writerSyncer, level)
			l.logger = zap.New(core, zap.AddCaller(), zap.Development())
			l.logger.Info("log initialized", zap.String("log location", filename), zap.String("level", level.String()))
		})

}

func (l *logUseCase) GetLogger() *zap.Logger {
	if l.logger != nil {
		return l.logger
	}
	l.logInit()
	return l.logger
}

func newWriter1(fileName string) io.Writer {
	lumberJackLogger := &lumberjack.Logger{
		LocalTime:  true,
		Filename:   fileName,
		MaxSize:    5,  //最大存储大小--MBytes
		MaxBackups: 10, //最大备份数量--Nums
		MaxAge:     5,  //最大存储时间--Days
		Compress:   true,
	}
	return lumberJackLogger
}

package logger

import (
	"encoding/json"
	"fmt"
	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"time"
)

var zapLog *zap.SugaredLogger

type Logger struct {
	KeyAndValues []interface{}
	LogLevel     string
	MsgChain     string
}

func NewLog() Logger {
	return Logger{
		LogLevel: "info",
	}
}
func (l *Logger) Append(msg string, keysAndValues ...interface{}) {
	l.KeyAndValues = append(l.KeyAndValues, keysAndValues...)
	if msg != "" {
		l.MsgChain = l.MsgChain + "|" + msg
	}
}
func (l *Logger) SetLogLevel(level string) {
	l.LogLevel = level
}

func (l *Logger) LogExceptInfo() {
	switch l.LogLevel { // 初始化配置文件的Level
	case "debug":
		zapLog.Debugw(l.MsgChain, l.KeyAndValues...)
	case "info":
		zapLog.Infow(l.MsgChain, l.KeyAndValues...)
	case "warn":
		zapLog.Warnw(l.MsgChain, l.KeyAndValues...)
	case "error":
		buf, _ := json.MarshalIndent(l, "  ", "  ")
		now := time.Now().String()
		buf = append(buf, now...)
		zapLog.Errorw(l.MsgChain, l.KeyAndValues...)
	case "dpanic":
		zapLog.DPanicw(l.MsgChain, l.KeyAndValues...)
	case "panic":
		zapLog.Panicw(l.MsgChain, l.KeyAndValues...)
	case "fatal":
		zapLog.Fatalw(l.MsgChain, l.KeyAndValues...)
		//default:
		//	zapLog.Debugw(l.MsgChain, l.KeyAndValues...)
	}
}
func (l *Logger) Log() {
	switch l.LogLevel { // 初始化配置文件的Level
	case "debug":
		zapLog.Debugw(l.MsgChain, l.KeyAndValues...)
	case "info":
		zapLog.Infow(l.MsgChain, l.KeyAndValues...)
	case "warn":
		zapLog.Warnw(l.MsgChain, l.KeyAndValues...)
	case "error":
		buf, _ := json.MarshalIndent(l, "  ", "  ")
		now := time.Now().String()
		buf = append(buf, now...)
		zapLog.Errorw(l.MsgChain, l.KeyAndValues...)
	case "dpanic":
		zapLog.DPanicw(l.MsgChain, l.KeyAndValues...)
	case "panic":
		zapLog.Panicw(l.MsgChain, l.KeyAndValues...)
	case "fatal":
		zapLog.Fatalw(l.MsgChain, l.KeyAndValues...)
	default:
		zapLog.Debugw(l.MsgChain, l.KeyAndValues...)
	}
}
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func InitZap(level string, srvName string, logPath string, alertWebHook string) {
	if ok, _ := pathExists(logPath); !ok { // 判断是否有Director文件夹
		_ = os.Mkdir(logPath, os.ModePerm)
	}
	var levelLimit zapcore.Level
	switch level { // 初始化配置文件的Level
	case "debug":
		levelLimit = zap.DebugLevel
	case "info":
		levelLimit = zap.InfoLevel
	case "warn":
		levelLimit = zap.WarnLevel
	case "error":
		levelLimit = zap.ErrorLevel
	case "dpanic":
		levelLimit = zap.DPanicLevel
	case "panic":
		levelLimit = zap.PanicLevel
	case "fatal":
		levelLimit = zap.FatalLevel
	default:
		levelLimit = zap.InfoLevel
	}
	//zap.AddStacktrace(zap.ErrorLevel)
	logger := zap.New(getEncoderCore(levelLimit, logPath))

	logger.WithOptions(zap.AddStacktrace(zap.ErrorLevel))
	if levelLimit == zap.DebugLevel {
		logger = logger.WithOptions(zap.AddStacktrace(zap.DebugLevel))
	}
	logger = logger.WithOptions(zap.Fields(zap.String("sname", srvName)))
	//logger = logger.WithOptions(zap.AddCaller())
	//logger = logger.WithOptions(zap.AddCallerSkip(1))

	zapLog = logger.Sugar()
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore(level zapcore.Level, logPath string) (core zapcore.Core) {
	writer, err := GetWriteSyncer(logPath) // 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(), writer, level)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
}

func GetWriteSyncer(p string) (zapcore.WriteSyncer, error) {
	fileWriter, err := zaprotatelogs.New(
		path.Join(p, "%Y-%m-%d.log"),
		//zaprotatelogs.WithLinkName("latest.log"),
		zaprotatelogs.WithMaxAge(3*24*time.Hour),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	if true {
		//if level == zap.DebugLevel {
		//	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
		//}
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}

func Debug(msg string, keysAndValues ...interface{}) {
	zapLog.Debugw(msg, keysAndValues...)
}
func Info(msg string, keysAndValues ...interface{}) {
	zapLog.Infow(msg, keysAndValues...)
}
func Warn(msg string, keysAndValues ...interface{}) {
	zapLog.Warnw(msg, keysAndValues...)
}
func Error(msg string, keysAndValues ...interface{}) {
	zapLog.Errorw(msg, keysAndValues...)
}
func ErrorSync(msg string, keysAndValues ...interface{}) {
	zapLog.Errorw(msg, keysAndValues...)
}
func Panic(msg string, keysAndValues ...interface{}) {
	zapLog.Panicw(msg, keysAndValues...)
}
func Fatal(msg string, keysAndValues ...interface{}) {
	zapLog.Fatalw(msg, keysAndValues...)
}

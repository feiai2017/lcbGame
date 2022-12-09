package main

import (
	"errors"
	_ "expvar"
	"flag"
	"fmt"
	"mania/apis"
	"mania/constant"
	"mania/control"
	"mania/logger"
	"mania/middleware"
	"mania/tcpx"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configPath  = flag.String("config", `config.json`, "config file path")
	showVersion = flag.Bool("version", false, "show server version")

	// BuildTime 编译参数
	BuildTime string
	CommitID  string
	LogLevel  string

	srv *tcpx.TcpX
)

func main() {
	t := time.Now()
	flag.Parse()

	if *showVersion {
		fmt.Printf("build-time:%s\ncommit-id:%s\nlog-level:%s\n", BuildTime, CommitID, LogLevel)
		os.Exit(0)
	}

	err := control.LoadConfigFromFile(*configPath)
	if err != nil {
		fmt.Printf("错误类型：%T\n错误信息：%v\n调用栈：\n%+v\n", errors.Unwrap(err), errors.Unwrap(err), err)
		return
	}

	err = control.Store.Connect(constant.ServerName)
	if err != nil {
		logger.ErrorSync("", "错误信息", errors.Unwrap(err), "调用栈：", err)
		return
	}
	defer control.Store.CloseDB()

	srv = tcpx.NewTcpX(tcpx.JsonMarshaller{})

	control.SrvConfig.BuildTime = BuildTime
	control.SrvConfig.CommitID = CommitID
	control.SrvConfig.LogLevel = LogLevel

	srv.OnClose = apis.OnClose
	srv.OnConnect = apis.OnConnect
	srv.SetMaxBytePerMessage(1024 * 1024)
	srv.Use("mutex", middleware.MutexConnection)
	srv.Use("logger", middleware.Logger)
	srv.Use("recover", middleware.Recover)
	srv.HeartBeatModeDetail(true, time.Minute, false, 200)
	srv.RewriteHeartBeatHandler(constant.HeartBeat, apis.HeartBeat)
	srv.AddHandler(constant.UserLogin, apis.UserLogin)
	srv.AddHandler(constant.SnakeLadder, apis.SnakeLadder)

	msg := fmt.Sprintf("build-time:%s\ncommit-id:%s\nlog-level:%s\ntime:%s\ncost:%dms\n",
		BuildTime, CommitID, LogLevel, time.Now().String(), time.Now().Sub(t).Milliseconds())

	logger.Warn("start", "msg", msg)
	fmt.Println("start")

	shutdown := make(chan int, 0)
	go DebugPProf()
	go Wait(shutdown)
	defer catch()

	err = srv.ListenAndServe("tcp", ":7170")
	if err != nil {
		logger.Warn("server启动失败：\n" + err.Error())
	}

	<-shutdown
}

func catch() {
	if r := recover(); r != nil {
		msg := fmt.Sprintf("build-time:%s\ncommit-id:%s\nlog-level:%s\nrecover:%s\ntime:%s\n",
			control.SrvConfig.BuildTime, control.SrvConfig.CommitID,
			control.SrvConfig.LogLevel, fmt.Sprintf("%v", r), time.Now().String())
		logger.Error("server关闭1", "msg", msg)
	}
}

// Wait 监听退出信号，等待处理完成
func Wait(shutdown chan int) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range signals {
			msg := fmt.Sprintf("build-time:%s\ncommit-id:%s\nlog-level:%s\nsignal:%s\ntime:%s\n",
				control.SrvConfig.BuildTime, control.SrvConfig.CommitID,
				control.SrvConfig.LogLevel, sig.String(), time.Now().String())
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				_ = srv.Stop(false)
				//if SaveAllAndCloseConn() {
				//}
				control.Store.CloseDB()
				logger.Warn("server关闭2", "msg", msg)
				close(shutdown)
			default:
				//logger.Error("获得其他信号", "msg", msg)
			}

		}
	}()
}

func DebugPProf() {
	if err := http.ListenAndServe(":6080", nil); err != nil {
		logger.Error("DebugPProf", "err", err)
	}
}

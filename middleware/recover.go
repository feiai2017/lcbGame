package middleware

import (
	"fmt"
	"mania/apis/protocol"
	"mania/constant"
	"mania/control"
	"mania/logger"
	"mania/tcpx"
	"runtime/debug"
	"time"
)

func Recover(c *tcpx.Context) {
	defer func() {
		if r := recover(); r != nil {
			rawUid, ok := c.GetCtxPerConn("uid")
			uid := 0
			if ok {
				uid = rawUid.(int)
			}
			var tu bool

			var duid string
			uuid, ok := c.GetCtxPerConn("uuid")
			if ok {
				duid = uuid.(string)
			}

			protocolID, _ := c.Packx.MessageIDOf(c.Stream)
			msg := &logger.PanicPayload{
				Happen:       time.Now(),
				Stack:        string(debug.Stack()),
				RemoteSocket: c.Conn.RemoteAddr().String(),
				LocalSocket:  control.SrvConfig.InnerIP,
				Recover:      fmt.Sprintf("%v", r),
				Protocol:     int(protocolID),
				UID:          uid,
				BuildTime:    control.SrvConfig.BuildTime,
				CommitID:     control.SrvConfig.CommitID,
				LogLevel:     control.SrvConfig.LogLevel,
				TestUser:     tu,
				ServerName:   control.SrvConfig.Name,
				Uuid:         duid,
			}

			rawLog, ok := c.GetCtxPerConn("logger")
			if ok {
				log := rawLog.(*logger.Logger)
				log.SetLogLevel("error")
				log.Append("panic", "key", "panic", "message", msg)
				//util.Warning(msg)
				log.Log()
			}

			err := &protocol.CommonError{
				Code: constant.ErrJsonMarshallerFail,
			}
			_ = c.JSON(protocolID, err)
			return
		}
	}()
	c.Next()
}

package apis

import (
	"mania/apis/protocol"
	"mania/constant"
	"mania/logger"
	"mania/model"
	"mania/service"
	"mania/tcpx"
	"time"
)

func SnakeLadder(c *tcpx.Context) {
	rawLog, _ := c.GetCtxPerConn("logger")
	log := rawLog.(*logger.Logger)
	log.Append("", "key", "snake")

	old := time.Now()
	var req protocol.ReqSnakeLadder
	var resp protocol.RespSnakeLadder

	defer func() {
		err := c.JSON(constant.SnakeLadder, resp)
		if err != nil {
			log.SetLogLevel("error")
			log.Append("send resp failed", "err", err)
		}
		now := time.Now()
		delta := now.Sub(old).Milliseconds()
		log.Append("", "delta", delta)
		log.LogExceptInfo()
	}()
	uid := 0
	rawUid, ok := c.GetCtxPerConn("uid")
	if !ok {
		log.SetLogLevel("warn")
		log.Append("uid is wrong")
		resp.Code = constant.ErrGetTcpUserFail
		return
	}
	uid = rawUid.(int)
	log.Append("", "uid", uid)
	usr, err := model.Get(uid)
	if err != nil {
		log.SetLogLevel("warn")
		log.Append("user is nil", "err", err)
		resp.Code = constant.ErrGetTcpUserFail
		return
	}
	usr.Log(log)
	_, err = c.BindWithMarshaller(&req, tcpx.JsonMarshaller{})
	if err != nil {
		log.SetLogLevel("error")
		log.Append("decode req failed", "err", err)
		resp.Code = constant.ErrJsonMarshallerFail
		return
	}
	log.Append("recv req", "req", req)

	switch req.Action {
	// 1:初始化棋盘 初始化N个队友
	case 1:
		service.InitSnakeLadderMap(usr)
	// 2:掷色子
	case 2:
		resp.Result, err = service.Roll(usr)
		if err != nil {
			log.SetLogLevel("error")
			log.Append("decode req failed", "err", err)
			resp.Code = constant.GameError
			return
		}
	// 3：回放
	case 3:
		if usr.SnakeLadder != nil && len(usr.SnakeLadder.SavePlayBack) > 0 {
			resp.Playback = usr.SnakeLadder.SavePlayBack
		}
	default:

	}

	resp.Code = constant.SUCCESS
	log.Append("", "resp", resp)
}

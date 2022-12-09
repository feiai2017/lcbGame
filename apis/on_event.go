package apis

import (
	"fmt"
	"mania/constant"
	"mania/control"
	"mania/logger"
	"mania/model"
	"mania/service"
	"mania/tcpx"
	"sync"
	"time"
)

func OnClose(c *tcpx.Context) {
	uid := 0
	rawUid, ok := c.GetCtxPerConn("uid")
	if !ok {
		return
	}

	log := logger.NewLog()
	log.Append("", "key", "on_close")

	old := time.Now()

	defer func() {
		now := time.Now()
		delta := now.Sub(old).Milliseconds()
		log.Append("", "delta", delta)
		log.Log()
	}()
	uid = rawUid.(int)

	log.Append("", "uid", uid)
	usr, err := model.Get(uid)
	if err != nil {
		log.SetLogLevel("warn")
		log.Append("user is nil", "err", err)
		return
	}

	usr.Log(&log)

	err2 := service.UpdateUserByObject(usr)
	if err2 != nil {
		// TODO: 重试
		log.SetLogLevel("error")
		log.Append(fmt.Sprintf("save user failed uid = [%d]", uid), "err2", err2)
	} else {
		err3 := control.Store.UserOffline(uid, constant.ServerName)
		if err3 != nil {
			log.SetLogLevel("error")
			log.Append("user offline failed", "err3", err3)
			return
		}

		model.Destroy(uid)
		c.PerConnectionContext.Delete("uid")
	}
}

func OnConnect(c *tcpx.Context) {
	mutex := new(sync.Mutex)
	c.SetCtxPerConn("mutex", mutex)
	c.SetCtxPerConn("open_socket", time.Now())
}

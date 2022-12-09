package apis

import (
	"go.mongodb.org/mongo-driver/bson"
	"mania/apis/protocol"
	"mania/constant"
	"mania/control"
	"mania/logger"
	"mania/model"
	"mania/service"
	"mania/tcpx"
	"time"
)

func UserLogin(c *tcpx.Context) {
	openSocketTime := time.Time{}
	openSocket, ok := c.GetCtxPerConn("open_socket")
	if ok {
		ost, ok := openSocket.(time.Time)
		if ok {
			openSocketTime = ost
		}
	}

	rawLog, _ := c.GetCtxPerConn("logger")
	log := rawLog.(*logger.Logger)
	log.Append("", "key", "user_login")

	if !openSocketTime.IsZero() {
		log.Append("", "socket_delta", time.Now().Sub(openSocketTime).Milliseconds(), "socket_open", openSocketTime.Unix())
	}

	old := time.Now()
	var req protocol.ReqUserLogin
	var resp protocol.RespUserLogin

	defer func() {
		err := c.JSON(constant.UserLogin, resp)
		if err != nil {
			log.SetLogLevel("error")
			log.Append("send resp failed", "err", err)
		}
		now := time.Now()
		delta := now.Sub(old).Milliseconds()
		log.Append("", "delta", delta)
		log.LogExceptInfo()
	}()
	_, err := c.BindWithMarshaller(&req, tcpx.JsonMarshaller{})
	if err != nil {
		log.SetLogLevel("error")
		log.Append("decode req failed", "err", err)
		resp.Code = constant.ErrJsonMarshallerFail
		return
	}
	log.Append("", "req", req)

	if req.Duid == "" {
		log.SetLogLevel("error")
		log.Append("err req argument")
		resp.Code = constant.ErrJsonMarshallerFail
		return
	}

	platform := 0
	duid := req.Duid

	uid, credential, usr, err, _ := service.Login(duid, platform, req.Duid)
	if err != nil {
		if err.Error() == "error credential" {
			log.SetLogLevel("warn")
			log.Append("user.Login failed", "err", err.Error())
			resp.Code = constant.ErrToken
			return
		}

		log.SetLogLevel("error")
		log.Append("user.Login failed", "err", err)
		resp.Code = constant.ErrJsonMarshallerFail
		return
	}
	if uid == 0 || credential == "" || usr == nil {
		log.SetLogLevel("error")
		log.Append("login or create user failed", "err", err)
		resp.Code = constant.ErrGetTcpUserFail
		return
	}
	log.Append("", "uid", uid)

	log.Append("Re login")
	// 获取用户数据
	oldUser, _ := model.Get(uid)
	if oldUser != nil {
		// 保存
		err = service.UpdateUserByObject(oldUser)
		if err != nil {
			log.SetLogLevel("error")
			log.Append("user.UpdateUserByObject", "err", err.Error())
		}

		// 如果是新的socket,关闭旧的socket
		oldTCPXContext := oldUser.Conn
		if oldTCPXContext != nil && oldTCPXContext.Conn != c.Conn {
			log.Append("oldTCPXContext.Conn != c.Conn")
			service.UserOut(oldTCPXContext, 3, true)
		}

		time.Sleep(time.Millisecond * 500)

		filter := bson.D{{Key: "uid", Value: uid}}
		usr, err = service.GetUser(filter)
		if err != nil {
			log.SetLogLevel("error")
			log.Append("user.GetUser error", "err", err.Error())
			resp.Code = constant.ErrGetTcpUserFail
			return
		}
	}

	_ = model.Set(usr)
	err = control.Store.UserOnline(usr.UID, constant.ServerName)
	if err != nil {
		log.SetLogLevel("error")
		log.Append("user online failed", "err", err)
		return
	}
	usr.Log(log)

	resp.Credential = credential
	resp.UserID = usr.UID

	usr.Package = req.Package
	usr.Conn = c
	c.SetCtxPerConn("uid", uid)

	log.Append("success")
	resp.Code = constant.SUCCESS
	logger.Debug("login debug", "uid", usr.UID, "req", req, "resp", resp)
}

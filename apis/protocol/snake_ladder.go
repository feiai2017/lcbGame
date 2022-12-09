package protocol

import "mania/model"

type ReqSnakeLadder struct {
	Action int `json:"action"` // 1:初始化棋盘 初始化N个队友 2:掷色子 3：回放
}

type RespSnakeLadder struct {
	CommonError
	Result   *model.RollSnakeLadder `json:"result,omitempty"`
	Playback []model.Points         `json:"playback,omitempty"`
}

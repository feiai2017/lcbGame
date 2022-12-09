package service

import (
	"fmt"
	"mania/control"
	"mania/model"
	"mania/util/rand"
)

const (
	MapLength = 100

	WinIndex = 99
)

// InitSnakeLadderMap 初始化地图
func InitSnakeLadderMap(usr *model.User) {
	// 获取用户在线数量
	users := control.Store.AllUserOnline()
	l := len(users)

	usr.SnakeLadder = &model.SnakeLadder{
		Map:       make([]int, MapLength),
		WinPlayer: 0,
	}
	temp := make(model.Points, l)
	for _, v := range users {
		temp[v] = -1
	}

	usr.SnakeLadder.Current = temp
	usr.SnakeLadder.SavePlayBack = make([]model.Points, l)

	usr.SnakeLadder.Number, _ = control.Store.CreateMapId()
	usr.SnakeLadder.Sequence = GenerateSequence(users)
}

// GenerateSequence 根据玩家生成序列
func GenerateSequence(users []int) (sequence []int) {
	sequence = make([]int, len(users))
	rand.Shuffle(users)
	for i := 0; i < len(users)*100; {
		for _, v := range users {
			sequence[i] = v
			i++
		}
	}

	return
}

func Roll(usr *model.User) (r *model.RollSnakeLadder, err error) {
	if usr.SnakeLadder == nil {
		err = fmt.Errorf("game not initialized")
		return
	}

	if usr.SnakeLadder.WinPlayer > 0 {
		err = fmt.Errorf("the game is over")
		return
	}

	// 校验顺序
	if usr.SnakeLadder.Sequence[0] != usr.UID {
		err = fmt.Errorf("wait for other players to roll")
		return
	}

	r = new(model.RollSnakeLadder)
	r.Points = make(model.Points, 1)

	rollPoint := rand.RandomInt(6) + 1
	r.Points[usr.UID] = rollPoint

	usr.SnakeLadder.Current[usr.UID] += rollPoint
	current := usr.SnakeLadder.Current[usr.UID]
	// 游戏结束
	if current == WinIndex {
		usr.SnakeLadder.WinPlayer = usr.UID
		r.SnakeLadder = usr.SnakeLadder
		err := control.Store.CreateMap(usr.SnakeLadder)
		if err != nil {
			return nil, err
		}
		// 回退
	} else if current > WinIndex {
		usr.SnakeLadder.Current[usr.UID] = WinIndex - (current - WinIndex)
	}

	// 保存回放
	usr.SnakeLadder.SavePlayBack = append(usr.SnakeLadder.SavePlayBack, r.Points)

	r.SnakeLadder = usr.SnakeLadder
	return
}

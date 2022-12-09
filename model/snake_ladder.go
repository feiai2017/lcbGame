package model

type SnakeLadder struct {
	Number       int      `json:"number"`         // 房间号
	Current      Points   `json:"current"`        // map[uid]index 从 -1 开始
	Map          []int    `json:"map"`            // 地图 ;0空
	SavePlayBack []Points `json:"save_play_back"` // 保存每次的骰子结果，用于回放
	WinPlayer    int      `json:"win_player"`     // 获胜人编号 0未分出胜负
	Sequence     []int    `json:"sequence"`       // 顺序
}

type RollSnakeLadder struct {
	Points      Points       `json:"points"`       // 玩家结果
	SnakeLadder *SnakeLadder `json:"snake_ladder"` // 位置 地图等信息
}

type Points map[int]int

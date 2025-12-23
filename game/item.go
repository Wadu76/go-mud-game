package game

type Item struct {
	//gorm.Model 包括了ID，CreatedAt，UpdatedAt，DeletedAt等字段
	//每件物品都有一个唯一的ID
	ID uint `gorm:"primaryKey" json:"id"`

	//每件物品的名字
	Name string `json:"name"`

	//每件物品都有其描述
	Desc string `json:"desc"`

	//外键，指向Player表。表示该item属于哪个Player
	PlayerName string `json:"player_name" gorm:"size:255"`

	//物品在玩家手里，Roomname为空，反之PlayerName为空
	//记录物品是否在房间里
	RoomName string `json:"room_name" gorm:"size:255"`
}

// 工厂函数
func NewItem(name string, desc string) *Item {
	return &Item{
		Name: name,
		Desc: desc,
	}
}

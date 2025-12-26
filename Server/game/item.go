package game

//物品有很多种 武器/杂物/药水道具等
const (
	ItemTypeGeneral = 0 //普通物品
	ItemTypeWeapon  = 1 //武器
	ItemTypePotion  = 2 //药水
)

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

	//物品类型
	Type int `json:"type"`

	//数值
	Value int `json:"value"`

	//是否装配标志量
	IsEquipped bool `json:"is_equipped"`
}

// 工厂函数
func NewItem(name string, desc string, itemType int, value int) *Item {
	return &Item{
		Name:       name,
		Desc:       desc,
		Type:       itemType,
		Value:      value,
		IsEquipped: false, //默认为未装配

	}
}

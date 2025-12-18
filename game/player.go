package game

import (
	"fmt"
	"mud-server/database"
)

// 1定义玩家结构体
// Capital == Public else Private
type Player struct {
	Name  string `gorm:"primaryKey" json:"name"` //玩家名字
	Level int    `json:"level"`                  //玩家等级
	HP    int    `json:"hp"`                     //玩家当前血量
	MaxHP int    `json:"max_hp"`                 //玩家最大血量

	//玩家所在房间 （类似个gps）
	CurrentRoom     *Room  `gorm:"-" json:"-"` //gorm: "-" json:"-" 表示在数据库中不存储，在json中也不展示，忽略
	CurrentRoomName string `json:"room_name"`

	//玩家背包 gorm会去item表中找 PlayerName == name的记录帮给他放入这个背包中
	Inventory []Item `gorm:"foreignKey:PlayerName" json:"inventory"`

}

// 从数据库加载玩家
func LoadPlayer(name string) (*Player, error) {
	var p Player

	//查玩家，顺便查他的背包
	result := database.DB.Preload("Inventory").Where("name = ?", name).First(&p)
	if result.Error != nil {
		//记录不存在说明是新玩家
		return nil, nil
	}
	return &p, nil
}

// 保存玩家到数据库 (只有Player类型能用)
func (p *Player) Save() error {
	//同步房间
	if p.CurrentRoom != nil {
		p.CurrentRoomName = p.CurrentRoom.Name
	}

	//写入
	result := database.DB.Save(p)

	if result.Error != nil {
		return result.Error
	}
	fmt.Printf(" 玩家[%s]数据已同步！\n", p.Name)
	return nil
}

//2定义玩家方法

func (p *Player) GetName() string {
	return p.Name
}

func (p *Player) TakeDamage(dmg int) string {
	p.HP -= dmg
	if p.HP < 0 {
		p.HP = 0
	}
	return fmt.Sprintf("  -> [%s] 受到了 %d 点伤害, 剩余HP %d/%d\n", p.Name, dmg, p.HP, p.MaxHP)
}

func (p *Player) Attack(target Attackable) string {
	damage := 10 //假设每次攻击造成10点伤害(暂时)
	log1 := fmt.Sprintf(" 🗡 [%s] 攻击了 [%s]!\n", p.Name, target.GetName())

	log2 := target.TakeDamage(damage)
	return log1 + "\n" + log2
}

func (p *Player) Heal() string {
	heal := 15 //规定每次恢复15血
	p.HP += heal
	if p.MaxHP < p.HP {
		p.HP = p.MaxHP
	}
	return fmt.Sprintf("💊 [%s] 治疗了自己，恢复 %d 点血量！目前血量为 %d\n", p.Name, heal, p.HP)
}

func NewPlayer(name string, level int, hp int, maxHp int) *Player {
	return &Player{
		Name:        name,
		Level:       level,
		HP:          hp,
		MaxHP:       maxHp,
		CurrentRoom: nil, //初始化时暂时为空，后面为World分配
	}
}

// 移动逻辑
func (p *Player) Move(direction string) (bool, string) {
	if p.CurrentRoom == nil {
		return false, "召唤师，你还在虚空中..."
	}

	//根据方向获取下一个房间
	nextRoom, ok := p.CurrentRoom.Exits[direction]
	if !ok {
		return false, "那边没有路！"
	}

	//玩家先离开该房间
	p.CurrentRoom.PlayerLeave(p)

	//移动，先把玩家对应房间信息更新
	p.CurrentRoom = nextRoom

	//玩家进入新房间
	p.CurrentRoom.PlayerEnter(p)

	return true, p.CurrentRoom.GetInfo()
}


//查看背包
func (p* Player) ListInventory() string {
	if (len(p.Inventory) == 0) {
		return "你的背包空空如也~ \n"
	}

	info := "你的背包里有：\n"
	for _, item := range p.Inventory {
		info += fmt.Sprintf("- [%s]: %s\n", item.Name, item.Desc)
	}
	return info
}

package game

import "fmt"

//1定义玩家结构体
//Capital == Public else Private
type Player struct {
	Name  string //玩家名字
	Level int    //玩家等级
	HP    int    //玩家当前血量
	MaxHP int    //玩家最大血量

	//玩家所在房间 （类似个gps）
	CurrentRoom *Room
}

//2定义玩家方法

func(p *Player) GetName() string {
	return p.Name
}

func(p *Player) TakeDamage(dmg int) string {
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

func (p *Player) Heal() string  {
	heal := 15 //规定每次恢复15血
	p.HP += heal
	if p.MaxHP < p.HP {
		p.HP = p.MaxHP
	}
	return fmt.Sprintf("💊 [%s] 治疗了自己，恢复 %d 点血量！目前血量为 %d\n", p.Name, heal, p.HP)
}

func NewPlayer(name string,level int, hp int, maxHp int) *Player {
	return &Player{
		Name:  name,
		Level: level,
		HP:    hp,
		MaxHP: maxHp,
		CurrentRoom: nil, //初始化时暂时为空，后面为World分配
	}
}

//移动逻辑
func (p *Player) Move(direction string) (bool, string) {
	if p.CurrentRoom == nil {
		return false, "召唤师，你还在虚空中..."
	}

	//根据方向获取下一个房间
	nextRoom, ok := p.CurrentRoom.Exits[direction]
	if !ok {
		return false, "那边没有路！"
	}

	//移动
	p.CurrentRoom = nextRoom
	return true, p.CurrentRoom.GetInfo()
}

package game

import "fmt"

//1定义玩家结构体
//Capital == Public else Private
type Player struct {
	Name  string //玩家名字
	Level int    //玩家等级
	HP    int    //玩家当前血量
	MaxHP int    //玩家最大血量
}

//2定义玩家方法

func(p *Player) GetName() string {
	return p.Name
}

func(p *Player) TakeDamage(dmg int) {
    p.HP -= dmg
	if p.HP < 0 {
		p.HP = 0
	}
	fmt.Printf("  -> [%s] 受到了 %d 点伤害, 剩余HP %d/%d\n", p.Name, dmg, p.HP, p.MaxHP)
}

func (p *Player) Attack(target Attackable) {
	damage := 10 //假设每次攻击造成10点伤害(暂时)
	fmt.Printf(" 🗡 [%s] 攻击了 [%s]!\n", p.Name, target.GetName())
	
	target.TakeDamage(damage)
}

func (p *Player) Heal(target *Player) {
	heal := 15 //规定每次恢复15血
	target.HP += heal
	if target.MaxHP < target.HP {
		target.HP = target.MaxHP
	}
	fmt.Printf("💊 [%s] 治疗了 [%s]，恢复 %d 点血量！目前血量为 %d\n", p.Name, target.Name, heal, target.HP)
}

func NewPlayer(name string,level int, hp int, maxHp int) *Player {
	return &Player{
		Name:  name,
		Level: level,
		HP:    hp,
		MaxHP: maxHp,
	}
}

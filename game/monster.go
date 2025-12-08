package game

import (
	"fmt"
)

type Monster struct {
	Name  string
	HP    int
	MaxHP int
	Exp   int //怪物凋落物
}

//必须实现GetName才能满足Attackable接口

func (m *Monster) GetName() string {
	return m.Name
}

// 同理要实现TakeDamage
func (m *Monster) TakeDamage(dmg int) {
	m.HP -= dmg
	if m.HP < 0 {
		m.HP = 0
	}
	fmt.Printf(" ->怪物 [%s] 受到 %d 点伤害，剩余HP %d/%d \n", m.Name, dmg, m.HP, m.MaxHP)
}

func (m *Monster) Attack(target Attackable) {
	damage := 10
	fmt.Printf(" ->怪物 [%s] 攻击了 [%s] \n", m.Name, target.GetName())
	target.TakeDamage(damage)
}
// 工厂函数
func NewMonster(name string, hp int, maxhp int, exp int) *Monster {
	return &Monster{
		Name:  name,
		HP:    hp,
		MaxHP: maxhp,
		Exp:   exp,
	}
}

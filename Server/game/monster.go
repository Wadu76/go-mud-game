package game

import (
	"fmt"
	"sync" //引入sync包,为了实现攻击怪物的并发安全。每次玩家攻击就要加锁，防止多个玩家同时攻击
	//当然要让monster先变成公共的怪物
)

type Monster struct {
	Name  string
	HP    int
	MaxHP int
	Exp   int //怪物掉落的经验
	sync.Mutex
	//有锁的玩家才能攻击怪物并使其扣血！

	//怪物的攻击力
	AttackVal int 
}

//必须实现GetName才能满足Attackable接口

func (m *Monster) GetName() string {
	return m.Name
}

// 同理要实现TakeDamage
func (m *Monster) TakeDamage(dmg int) string {
	m.Lock()
	defer m.Unlock() //函数结束后解锁

	m.HP -= dmg
	if m.HP < 0 {
		m.HP = 0
	}
	return fmt.Sprintf(" ->怪物 [%s] 受到 %d 点伤害，剩余HP %d/%d \n", m.Name, dmg, m.HP, m.MaxHP)
	
}

//attack不需要上锁，attack只用读，不用写
func (m *Monster) Attack(target Attackable) string {
	damage := m.AttackVal
	log1 := fmt.Sprintf(" ->怪物 [%s] 攻击了 [%s] \n", m.Name, target.GetName())
	log2 := target.TakeDamage(damage)

	return log1 + "\n" + log2 //返回两行拼起来的日志

}
// 工厂函数
func NewMonster(name string, hp int, maxhp int, exp int, attack int) *Monster {
	return &Monster{
		Name:  name,
		HP:    hp,
		MaxHP: maxhp,
		Exp:   exp,
		AttackVal: attack,
	}
}

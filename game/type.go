package game

//接口，定义“行为”， 不定义“数据”
//只要实现了 Name() 和 TakeDamage()的都是attackable的
type Attackable interface {
	GetName() string
	TakeDamage(dmg int)
}
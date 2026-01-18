package game

import (
	"encoding/json"
	"fmt"
	"mud-server/database"
)

// 定义一个发给客户端的结构体
type ItemDTO struct {
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Value      int    `json:"value"`
	IsEquipped bool   `json:"is_Equipped"`
}

// 定义玩家结构体
// Capital == Public， else Private
type Player struct {
	Name     string `gorm:"primaryKey" json:"name"` //玩家名字
	Password string `json:"-"`                      //-表示表示以后把玩家数据发给前端时，不要把密码也发过去（安全！）
	Level    int    `json:"level"`                  //玩家等级
	HP       int    `json:"hp"`                     //玩家当前血量
	MaxHP    int    `json:"max_hp"`                 //玩家最大血量

	//玩家所在房间 （类似个gps）
	CurrentRoom     *Room  `gorm:"-" json:"-"` //gorm: "-" json:"-" 表示在数据库中不存储，在json中也不展示，忽略
	CurrentRoomName string `json:"room_name"`

	//玩家背包 gorm会去item表中找 PlayerName == name的记录帮给他放入这个背包中
	Inventory []Item `gorm:"foreignKey:PlayerName" json:"inventory"`

	//玩家需要升级，玩家可以变强了
	Exp int `json:"exp"`

	//玩家升下一级需要的经验值
	NextLevelExp int `json:"next_level_exp"`
}

func NewPlayer(name string, level int, hp int, maxHp int) *Player {
	return &Player{
		Name:         name,
		Level:        level,
		HP:           hp,
		MaxHP:        maxHp,
		CurrentRoom:  nil, //初始化时暂时为空，后面为World分配
		Exp:          0,   //初始经验从0开始
		NextLevelExp: 100, //初始升级经验值 从1级->2级需要100经验
	}
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
	return fmt.Sprintf("  -> [%s] 受到了 %d 点伤害, 剩余HP %d/%d\n|CMD:HP:%s:%d:%d", p.Name, dmg, p.HP, p.MaxHP, p.Name, p.HP, p.MaxHP)
}

func (p *Player) Attack(target Attackable) string {
	//damage := 10 //假设每次攻击造成10点伤害(暂时) 我们已经有我们的数值计算函数了！
	damage := p.GetAttackPower()
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

// 查看背包
// 加上查看到已经安装装备
func (p *Player) ListInventory() string {
	if len(p.Inventory) == 0 {
		return "你的背包空空如也~ \n"
	}

	info := "你的背包里有：\n"
	for _, item := range p.Inventory {
		info += fmt.Sprintf("- [%s]: %s\n", item.Name, item.Desc)
	}

	for _, item := range p.Inventory {
		status := ""
		if item.IsEquipped {
			status = " (已装备)"
		}
		info += fmt.Sprintf("- %s[%s] (攻:%d)%s: %s\n", status, item.Name, item.Value, status, item.Desc)
		info += fmt.Sprintf("总攻击力:%d\n", p.GetAttackPower())
	}
	return info
}

// 丢弃物品 从背包inventory -> 地上 Drop item_name
func (p *Player) Drop(itemName string) (bool, string) {
	//先看有没有这个东西
	var targetItem *Item
	var index int
	for i, item := range p.Inventory {
		if item.Name == itemName {
			targetItem = &p.Inventory[i] //指针，指向背包里的物品
			index = i
			break
		}
	}
	if targetItem == nil {
		return false, "你背包里没有该物品诶"
	}

	//更新数据库,把物品拿下来，也就是把Item对应的PlayerName置空；
	//让Item对应的RoomName更新为丢弃的房间名
	//先置空Item对应的PlayerName
	targetItem.PlayerName = nil
	//再更新Item对应的RoomName(即player当前房间名字)
	targetItem.RoomName = p.CurrentRoom.Name

	if err := database.DB.Save(targetItem).Error; err != nil {
		return false, "数据库更新失败,丢不掉了QAQ" + err.Error()
	}

	//从玩家背包里删除 （删除切片元素，套模板写法）  ...用于展开切片用作分开的参数
	//:index表示index前的所有元素，为第一个参数；
	//index+1:表示index后的所有元素，为第二个参数
	//...将切片展开作为多个参数，也就是把上述俩作为参数，做到了删除index这一元素的作用
	p.Inventory = append(p.Inventory[:index], p.Inventory[index+1:]...)

	//别忘了房间Room还有个Item列表，表示每个房间里有哪些物品，所以也得让房间知道自己地上有哪些物品
	//Room的Itemmap key为Item.Name, value为*Item
	p.CurrentRoom.Items[targetItem.Name] = targetItem

	//前面的都没返回说明丢弃成功了
	return true, fmt.Sprintf("你丢弃了%s", targetItem.Name)

}

// 捡东西 从地图房间 -> 背包 pick itemName
func (p *Player) Pick(itemName string) (bool, string) {
	//依旧先检查地上有没有这个物品
	targetItem, ok := p.CurrentRoom.Items[itemName]
	if !ok {
		return false, "地上没这个玩意诶"
	}

	//更新数据库，把Item对应的PlayerName更新为当前玩家名字
	targetItem.PlayerName = &p.Name
	//再更新Item对应的RoomName为空,这样就能表示为地上的东西被玩家pick起来了
	targetItem.RoomName = ""

	if err := database.DB.Save(targetItem).Error; err != nil {
		return false, "数据库更新失败,捡不起来了QAQ"
	}

	//从房间中移除这个物品，别忘了Room有Itemmap！！！否则一直捡起来了
	delete(p.CurrentRoom.Items, itemName) //第一个参数是map，第二个是map的key

	//把捡起来的物品加入玩家背包
	p.Inventory = append(p.Inventory, *targetItem)

	//前面的都没返回说明捡起来了
	return true, fmt.Sprintf("你捡起了%s", targetItem.Name)
}

// 装配武器方法，穿戴上装备才有用！
func (p *Player) Equip(itemName string) (bool, string) {
	//先检查包里面有没有
	var targetItem *Item
	msg := ""
	//这个循环中的item拷贝了背包物品，值拷贝无法修改原物品
	// （_表示不要索引，但第二种循环表明了我们还是需要的） go语言的存在即合理
	/*for _, item := range p.Inventory {
		if item.Name == itemName {
			targetItem = &item
			break
		}
	}*/

	//而该循环方式我们是直接调用Inventory对应的物品，引用直接改

	for i := range p.Inventory {
		if p.Inventory[i].Name == itemName {
			targetItem = &p.Inventory[i]
			break
		}
	}
	if targetItem == nil {
		return false, "你背包里没有这个装备诶"
	}

	//检查类型是否为武器
	if targetItem.Type != ItemTypeWeapon {
		return false, "这个不是武器诶, 不能拿照片砍人吧！"
	}

	//检查是否已经装备了武器
	if targetItem.IsEquipped {
		return false, "你已经在装备这个武器了"
	}
	//不过这个有点歧义，我从背包里拿出来了，武器还在背包里吗？后面再改吧，先检测是否装备了？

	//如果已经拿了别的武器，要先卸下,

	/*
		if p.EquipedWeapon != nil {
			return false, "你已经在装备别的武器了，请先卸下"
		}*/

	// 额但是没弄player对应手上武器标签，后续更新,暂时先注释大概写法

	for i := range p.Inventory {
		//如果是武器，且已装备，且不是我现在要穿的这把
		if p.Inventory[i].Type == ItemTypeWeapon && p.Inventory[i].IsEquipped && p.Inventory[i].ID != targetItem.ID {
			p.Inventory[i].IsEquipped = false
			database.DB.Save(&p.Inventory[i]) // 记得存库
			msg = fmt.Sprintf("为你卸下了%s", p.Inventory[i].Name)
			//return true, fmt.Sprintf("安装上了%s, 且同时为你卸下了%s\n你现在攻击力为%d", targetItem.Name, p.Inventory[i].Name, p.GetAttackPower())
		}
	}
	//装备上武器
	targetItem.IsEquipped = true

	//存入数据库
	database.DB.Save(targetItem)

	msg += fmt.Sprintf("你装上了了%s 攻击力提升%d！", targetItem.Name, targetItem.Value)
	return true, msg
}

// 卸下武器方法
func (p *Player) Unequip(itemName string) (bool, string) {
	//先检查
	var targetItem *Item
	//先在背包里找这个武器吧，要是player有个装备标签的话，那应该更好
	/*for _, item := range p.Inventory {
		if item.Name == itemName {
			targetItem = &item
			break
		}
	}*/
	for i := range p.Inventory {
		if p.Inventory[i].Name == itemName {
			targetItem = &p.Inventory[i]
			break
		}
	}

	//压根没有该武器
	if targetItem == nil {
		return false, "你貌似没有佩戴任何武器（背包里没有）"
	}

	//不是武器
	if targetItem.Type != ItemTypeWeapon {
		return false, "这个不是武器诶"
	}

	//已经卸下了
	if !targetItem.IsEquipped {
		return false, "你已经卸下了这个武器"
	}

	//卸下武器
	targetItem.IsEquipped = false

	//存入数据库 卸下了~
	database.DB.Save(targetItem)

	return true, fmt.Sprintf("你卸下了%s, 攻击力减少了%d", targetItem.Name, targetItem.Value)
}

// 计算攻击力，让安装上武器有伤害
func (p *Player) GetAttackPower() int {
	//基础的拳头伤害
	damage := 1

	//遍历，把装备上的武器的伤害加起来
	for _, item := range p.Inventory {
		if item.IsEquipped {
			damage += item.Value
		}
	}

	//返回总共的伤害
	return damage
}

// 获取经验方法
func (p *Player) GainExp(amount int) string {
	p.Exp += amount
	log := fmt.Sprintf("你获得了%d点经验", amount)

	//检查是否升级
	//可能一次升级多次，因此用循环
	for p.Exp >= p.NextLevelExp {
		p.Level++
		p.Exp -= p.NextLevelExp
		p.NextLevelExp = p.Level * 100 //升级曲线，每级多100点，先这样简单啦

		//升级属性提高！
		p.MaxHP += 20
		//升级直接回满血
		p.HP = p.MaxHP

		log += fmt.Sprintf("\n你升级了！当前等级：%d", p.Level)

	}
	database.DB.Save(p)

	return log
}

// 获取背包数据的协议字符串
func (p *Player) GetInventoryProtocol() string {
	var dtos []ItemDTO
	for _, item := range p.Inventory {
		dtos = append(dtos, ItemDTO{
			Name:       item.Name,
			Desc:       item.Desc,
			Value:      item.Value,
			IsEquipped: item.IsEquipped,
		})
	}

	//转为json
	//此处_表示忽略返回值 err
	jsonData, _ := json.Marshal(dtos)
	//拼凑协议头 |CMD:INC:json
	return "|CMD:INC:" + string(jsonData)
}

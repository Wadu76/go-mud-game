package network

import (
	//"bufio"
	"fmt"
	"mud-server/ai"
	"mud-server/database"
	"mud-server/game"
	"net"
	"strings"
)

// StartServer 启动TCP服务器
func StartServer() {
	//0 先连数据库
	database.InitDB()
	//0-1 自动建表，根据game.player结构创建表
	//0-2 新加个Item表
	database.DB.AutoMigrate(&game.Player{}, &game.Item{})

	fmt.Println("正在检查并自动建表...")
	err := database.DB.AutoMigrate(&game.Player{}, &game.Item{})
	if err != nil {

		panic("自动建表失败: " + err.Error())
	}
	fmt.Println("表结构同步完成！")

	InitWorld()
	//1监听端口 8888
	listener, err := net.Listen("tcp", ":8888") //err是错误信息， listener是监听对象
	if err != nil {
		fmt.Println("启动服务器失败：", err)
		return
	}
	//defer确保函数退出前关闭listener
	defer listener.Close()

	fmt.Println(" 🚀游戏服务已启动，正在监听8888端口...")

	//2等待客户端连接，无限循环a
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("连接建立失败:", err)
			continue
		}
		fmt.Println(" 有新玩家连接:", conn.RemoteAddr())

		//3开启一个Goroutine处理新玩家
		go handleConnection(conn) //将handleConnection 即 放到另一个协程中处理
	}
}

// 处理单个玩家的连接
func handleConnection(conn net.Conn) {

	//验证，只有登陆成功才能跳出该训话
	var hero *game.Player

	conn.Write([]byte("欢迎来到瓦度世界！请输入你的名字：\n"))
	buf := make([]byte, 1024)
	for {
		//读取客户端输入的名字,
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		//去掉名字中的空格 \r\n  不然输入名字的时候你的名字将会是 名字 \r\n，啥都不输入也会叫\r\n
		input := strings.TrimSpace(string(buf[:n]))
		//将输入的字符串按空格分割成数组,为了识别名字 & 密码
		parts := strings.Fields(input)

		if len(parts) < 3 {
			conn.Write([]byte("格式错误，请输入: LOGIN <名字> <密码> 或 REGISTER <名字> <密码>\n>"))
			continue
		}

		//cmd是命令 LOGIN REGISTER 登录/注册 为了统一，我们统一大写
		cmd := strings.ToUpper(parts[0])

		//name是用户名
		name := parts[1]

		//pwd是密码
		pwd := parts[2]

		//注册逻辑
		if cmd == "REGISTER" {
			//先查看有没有重名的
			exists, _ := game.LoadPlayer(name)
			if exists != nil {
				conn.Write([]byte("该名字已经被注册了，请换一个\n"))
				continue
			}

			//创建新号
			//新号的等级是1，血量是100，默认的
			newHero := game.NewPlayer(name, 1, 100, 100)
			//密码绑定，
			newHero.Password = pwd

			//存入数据库中
			if err := database.DB.Create(newHero).Error; err != nil {
				conn.Write([]byte("注册失败，数据库出错\n"))
				continue
			}
			conn.Write([]byte("注册成功！请使用LOGIN登陆吧！\n"))

		} else if cmd == "LOGIN" {
			//登录逻辑
			//先确认有没有这个人
			loadedHero, err := game.LoadPlayer(name)

			//如果数据库找不到这个人，说明还没注册过
			if err != nil || loadedHero == nil {
				conn.Write([]byte("该用户不存在, 先去注册吧~\n"))
				continue
			}

			//校验密码对不对 （用户账号安全）
			if loadedHero.Password != pwd {
				conn.Write([]byte("密码错误，请重新输入\n"))
				continue
			}

			//没错就对了，登陆成功咯
			//既然是登陆的，那就不需要重新创建了，直接用数据库中的数据就行了
			hero = loadedHero
			//conn.Write([]byte("登陆成功！欢迎回来,%s \n", hero.Name))
			conn.Write([]byte(fmt.Sprintf("登录成功！欢迎回来，%s (Lv.%d)\n", hero.Name, hero.Level)))

			//这一行代码会发给客户端，客户端拦截后会初始化底部的血条。刚进入游戏的玩家就能根据自己当前血量显示血条了
			conn.Write([]byte(fmt.Sprintf("|CMD:HP:%s:%d:%d", hero.Name, hero.HP, hero.MaxHP)))

			break
			//登陆成功就可以跳出循环了，这一循环就是为了保障玩家账户安全
		} else {
			//如果既不是注册也不是登陆，那就说明输入的指令不对
			conn.Write([]byte("指令错误，请重新输入	LOGIN / REGISTER\n"))
			continue
		}

		//既然我们要弄密码了，那就不能弄一样的默认名了，就不弄了
	}
	/*n, err := conn.Read(buf) //n是读取到的字节数
	if err != nil {
		return
	}

	//起名字
	playername := string(buf[:n]) //buf[0 - n-1]
	//去掉名字中的空格 \r\n  不然输入名字的时候你的名字将会是 名字 \r\n，啥都不输入也会叫\r\n
	playername = strings.TrimSpace(playername)

	if playername == "" {
		playername = "不起名字（香菜版）"
	}

	fmt.Printf("正在尝试加载玩家:%s ...\n", playername)
	hero, err := game.LoadPlayer(playername)
	if err != nil {
		fmt.Println("读取失败")
		return
	}

	if hero != nil {
		fmt.Printf("欢迎回来老朋友，找个位置随便坐吧%s！ (等级 %d)\n", hero.Name, hero.Level)
		hero.CurrentRoom = GlobalWorld.StartRoom

		conn.Write([]byte(fmt.Sprintf("欢迎回来, %s!读取档案成功。\n", hero.Name)))
	} else {
		fmt.Printf("创建新角色: %s\n", playername)
		hero = game.NewPlayer(playername, 1, 100, 100)
		hero.CurrentRoom = GlobalWorld.StartRoom

		database.DB.Create(hero) //创建玩家
		conn.Write([]byte("欢迎你！你的数据已存储！\n"))
	}*/
	//初始化玩家位置 必须要弄，不然又要空指针报错了
	//如果玩家上次下线有记录房间，就去那个房间；如果没有（或找不到），就去新手村。
	if targetRoom, ok := GlobalWorld.AllRooms[hero.CurrentRoomName]; ok {
		hero.CurrentRoom = targetRoom
	} else {
		//如果是新号，或者上次的房间名字读不出来，就去出生点
		hero.CurrentRoom = GlobalWorld.StartRoom
	}

	//初始化玩家游戏数据
	//hero := game.NewPlayer(playername, 1, 100, 100)
	//monster := game.NewMonster("史莱姆王", 50, 50, 20)
	//此处正式把玩家丢到出生点
	//hero.CurrentRoom = GlobalWorld.StartRoom

	//加入世界 先加入世界World，再加入Room
	GlobalWorld.AddPlayer(hero.Name, conn)
	hero.CurrentRoom.PlayerEnter(hero)
	//defer conn.Close() //玩家断开时关闭连接

	//测试代码，先每个人发一把剑 测试成功，已经完成背包雏形，但目前还不能对背包进行操作
	if len(hero.Inventory) == 0 {
		sword := game.NewItem("破旧的铁剑", "工匠奥利弗打造的,不过现在有些破旧了", game.ItemTypeWeapon, 5)
		sword.PlayerName = &hero.Name

		database.DB.Create(sword)

		//更新背包
		hero.Inventory = append(hero.Inventory, *sword)
		fmt.Println("默认武器已发放")

	}

	defer func() {
		fmt.Println("saving...")
		hero.Save() //退出自动保存
		GlobalWorld.RemovePlayer(hero.Name, conn)
		//玩家退出后自动离开该房间，到时候回来依旧在此房间，因为在哪是和玩家绑定的，这里解绑的是Room里存的玩家信息
		hero.CurrentRoom.PlayerLeave(hero)
		conn.Close()

	}()

	fmt.Println("新玩家接入，正在初始化游戏数据...")
	GlobalWorld.MessageChannel <- fmt.Sprintf("欢迎 勇士 [%s] 加入游戏！\n", hero.Name)

	conn.Write([]byte("===== 欢迎来到GO MUD 在线测试版 =====\n 请输入 attack, heal, status, say, go, look, inventory, pick, drop, equip, unequip, save, exit\n>"))
	//Write 是一个核心方法，它的作用是将数据写入到一个“目标”中。 可以是文件、网络连接、内存缓冲区、标准输出（你的终端屏幕）等等。
	buf = make([]byte, 1024) //缓冲区
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("玩家断开连接:", conn.RemoteAddr())
			return
		}

		//去掉空格和换行
		input := string(buf[:n])
		line := strings.TrimSpace(input)

		//处理空指令
		if line == "" {
			conn.Write([]byte("> "))
			continue
		}

		//智能切割：把 "say hello world" 切成 ["say", "hello", "world"]
		//Fields 是一个核心方法，它的作用是将字符串按照指定的分隔符进行切割，返回一个字符串切片。
		//parts := strings.Fields(input) input没有清除空格，后面在verb种parts[0]若空（比如输入\n) 数组会越界报错！
		parts := strings.Fields(line)

		verb := strings.ToLower(parts[0])

		//游戏逻辑路由
		var response string //f发回给客户端的话

		boss := GlobalWorld.Boss

		switch verb {
		case "attack":
			log1 := hero.Attack(boss)
			/*给自己放个假，晚点改
			//血量更新的json msg
			//这里是玩家攻击boss，所以boss是target，对应数据就该填boss的
			//但要在这里改还是在player.go里改？？
			hpMsg := game.HPUpdateMsg{
				TargetName: boss.Name,
				CurrentHP: boss.HP,
				MaxHP: boss.MaxHP,
			}

			//包装一下
			serverMsg := game.ServerMessage{
				//这个事件是血量变化事件，告诉unity
				Event: "hp_change",
				//血量变化的数据
				Data: hpMsg,
			} */

			//广播给所有玩家，替换原本的response
			boradcastMsg := fmt.Sprintf("%s\n", log1)
			//GlobalWorld.MessageChannel <- boradcastMsg
			//response = log1 + "\n" 这是给单独玩家的response
			GlobalWorld.BroadcastToRoom(hero.CurrentRoom, boradcastMsg)

			if boss.HP > 0 {
				//boss反击
				log := boss.Attack(hero)

				response += log + "\n"
			} else {
				//boss被击败，肯定要广播
				GlobalWorld.MessageChannel <- fmt.Sprintf("勇士 [%s]成功击败了史莱姆王！获得 %d 经验\n", hero.Name, boss.Exp)
				//response += fmt.Sprintf("成功击败了史莱姆王！获得 %d 经验\n", boss.Exp)

				//发经验
				levelUplog := hero.GainExp(boss.Exp)
				conn.Write([]byte(levelUplog + "\n"))

				//重生boss

				boss.Lock()
				boss.HP = boss.MaxHP
				boss.Unlock()
				GlobalWorld.BroadcastToRoom(hero.CurrentRoom, fmt.Sprintf(" [%s] 复活了！快跑啊！\n", boss.Name))

			}

		case "heal":
			log1 := hero.Heal()
			//治疗，有破绽就被攻击了，目前治疗只能治疗自己。
			log2 := boss.Attack(hero)

			response = log1 + "\n" + log2 + "\n"

		case "status":
			//太宣布事故Boss全局的状态血量
			response = fmt.Sprintf("状态：[%s] HP: %d/%d VS [%s] HP: %d/%d\n", hero.Name, hero.HP, hero.MaxHP, boss.Name, boss.HP, boss.MaxHP)

		case "say":
			if len(parts) < 2 {
				response = "格式错误，say <内容>\n"
				break
			}

			content := line[len(parts[0]):]

			content = strings.TrimSpace(content)

			msg := fmt.Sprintf("[%s]说 %s\n", hero.Name, content)
			//GlobalWorld.MessageChannel <- msg
			GlobalWorld.BroadcastToRoom(hero.CurrentRoom, msg)
			response = ""

		case "go":
			if len(parts) < 2 {
				response = "要去哪？请输入 go north/south/east/west\n"
				break
			}
			direction := strings.ToLower(parts[1]) //提取第二个参数 即方向并将其改为小写
			//Move方法接受的是north/south/east/west 而不是中文，起始把direction改成中文传入Move中导致一直move不了
			success, info := hero.Move(direction)
			switch direction {
			case "north":
				direction = "北"
			case "south":
				direction = "南"
			case "east":
				direction = "东"
			case "west":
				direction = "西"
			}

			if success {
				response = fmt.Sprintf("你将向 %s ,进入 %s...\n ", direction, info)
			} else {
				//如果移动失败，则返回失败信息,在move里已经处理了走不通的报错逻辑
				response = info + "\n"
			}

		case "look":
			if hero.CurrentRoom == nil {
				response = hero.CurrentRoom.GetInfo() + "\n" //getinfo 里已经处理了空房间的情况
			} else {
				response = hero.CurrentRoom.GetInfo() + "\n"
			}
			//其实不需要，因为已经处理了空房间的情况，但为方便阅读就这样写了

		case "inventory":
			//response = hero.ListInventory()
			invMsg := hero.GetInventoryProtocol()
			conn.Write([]byte(invMsg))
			continue
			//直接跳过最后的conn.Write([]byte(response)),因为已经发送了协议

		//pick itemName
		case "pick":
			if len(parts) < 2 {
				response = "要捡什么？请输入 pick <物品名>\n"
				break
			}
			itemName := parts[1] //提取第二个参数 即物品名(不能有空格)
			ok, msg := hero.Pick(itemName)
			response = msg + "\n"
			if ok {
				GlobalWorld.BroadcastToRoom(hero.CurrentRoom, fmt.Sprintf("%s 捡起了 [%s]\n", hero.Name, itemName))
			}

		//drop itemName
		case "drop":
			if len(parts) < 2 {
				response = "要丢弃什么？请输入 drop <物品名>\n"
				break
			}
			itemName := parts[1] //提取第二个参数 即物品名(不能有空格)
			ok, msg := hero.Drop(itemName)
			response = msg + "\n"
			if ok {
				GlobalWorld.BroadcastToRoom(hero.CurrentRoom, fmt.Sprintf("%s 丢弃了 [%s]\n", hero.Name, itemName))
			}

		case "equip":
			if len(parts) < 2 {
				response = "要装备什么？请输入 equip <物品名>\n"
				break
			}
			itemName := parts[1] //提取第二个参数 即物品名(不能有空格)
			ok, msg := hero.Equip(itemName)
			response = msg + "\n"
			if ok {
			}
			//GlobalWorld.BroadcastToRoom(hero.CurrentRoom, fmt.Sprintf("%s 装备了 [%s]\n", hero.Name, itemName))

		case "unequip":
			if len(parts) < 2 {
				response = "要卸下什么？请输入 unequip <物品名>\n"
				break
			}
			itemName := parts[1] //提取第二个参数 即物品名(不能有空格)
			ok, msg := hero.Unequip(itemName)
			response = msg + "\n"
			if ok {
			}

			// 请把这段代码加到 switch cmd { ... } 里面

		case "talk":
			//指令格式: talk 守卫 你好啊
			if len(parts) < 3 {
				conn.Write([]byte("格式错误，请使用: talk <NPC名字> <想说的话>\n"))
				continue
			}

			targetName := parts[1]
			//把剩下的部分拼起来作为对话内容
			content := strings.Join(parts[2:], " ")

			//简单的 NPC 查找逻辑 (为了演示，我们硬编码一个守卫)
			//实际项目中这里会去 Room 里查找有没有这个NPC
			if targetName == "守卫" || targetName == "guard" {
				conn.Write([]byte(fmt.Sprintf("你对 [守卫] 说: %s\n", content)))

				//先给玩家一个反馈，让他知道 AI 正在思考
				conn.Write([]byte("Wait [守卫] 正在打量你...\n"))

				//开启协程异步请求AI
				//这样主线程不会被阻塞，其他玩家完全感觉不到卡顿
				go func(c net.Conn, playerMsg string) {
					//定义守卫的人设 (Persona)
					persona := "你是一个身经百战的皇家守卫，负责看守新手村大门。你性格傲慢，看不起衣衫褴褛的新手，说话喜欢带刺，但职责所在会回答关于怪物的问题。"

					//请求Kimi
					reply := ai.AskNPC("守卫", persona, playerMsg)

					//拿到结果，推发给客户端
					//注意格式：加个换行和颜色让它显眼一点
					c.Write([]byte(fmt.Sprintf("\n[守卫] 居然回复了: %s\n> ", reply)))
				}(conn, content)

			} else {
				conn.Write([]byte("这里没有这个人。\n"))
			}

		case "save":
			response = "保存成功\n"
			hero.Save()

		case "exit":
			conn.Write([]byte("Bye~\n"))

		default:
			response = fmt.Sprintf("未知指令 '%s'，请输入 attack, heal, status\n", verb)
		}

		if hero.HP <= 0 {
			response += "/(ㄒoㄒ)/~~ 胜败乃兵家常事，重新连接复活再来吧！\n"
			conn.Write([]byte(response))
			//return //踢走输掉的玩家
			//若是直接踢走会导致客户端展现为啥都按不了了，只能关掉重来？
			//那么为了形成一个游戏的闭环，我们就让角色直接在初始出生点复活
			//考虑要不要丢掉背包里的所有东西，先写个丢掉所有东西的大概逻辑，后面再考虑完善或考虑用不用
			/* ...此处省略丢掉所有东西的代码
			可以直接用Drop()函数，遍历整个背包全都丢掉！
			for _, item := range hero.Inventory {
				hero.Drop(item.Name)
				}
			*/

			//数据库中清空背包
			database.DB.Where("player_name =?", hero.Name).Delete(&game.Item{})

			//内存中清空背包
			hero.Inventory = []game.Item{}
			conn.Write([]byte("背包已被清空\n"))

			//复活
			hero.HP = hero.MaxHP

			//传送回出生点
			hero.CurrentRoom = GlobalWorld.StartRoom

			//告诉Unity自己的血量，用于更新自己的血条
			conn.Write([]byte(fmt.Sprintf("|CMD:HP:%s:%d:%d\n", hero.Name, hero.HP, hero.MaxHP)))

			conn.Write([]byte("欢迎回家，重走来时路吧！\n>"))
			continue
		}
		//最终战斗信息
		response += ">"
		conn.Write([]byte(response))

	}

}

package network

import (
	//"bufio"
	"fmt"
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

	conn.Write([]byte("欢迎来到瓦度世界！请输入你的名字：\n"))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) //n是读取到的字节数
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
	}

	//测试代码，先每个人发一把剑 测试成功，已经完成背包雏形，但目前还不能对背包进行操作
	if len(hero.Inventory) == 0 {
		sword := game.NewItem("破旧的铁剑", "工匠奥利弗打造的,不过现在有些破旧了", game.ItemTypeWeapon, 5)
		sword.PlayerName = hero.Name

		database.DB.Create(sword)

		//更新背包
		hero.Inventory = append(hero.Inventory, *sword)
		fmt.Println("默认武器已发放")

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

	defer func() {
		fmt.Println("saving...")
		hero.Save() //退出自动保存
		GlobalWorld.RemovePlayer(hero.Name, conn)
		//玩家退出后自动离开该房间，到时候回来依旧在此房间，因为在哪是和玩家绑定的，这里解绑的是Room里存的玩家信息
		hero.CurrentRoom.PlayerLeave(hero)
		conn.Close()

	}()

	fmt.Println("新玩家接入，正在初始化游戏数据...")
	GlobalWorld.MessageChannel <- fmt.Sprintf("欢迎 勇士 [%s] 加入游戏！\n", playername)

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
			}

		case "heal":
			log1 := hero.Heal()
			//治疗，有破绽就被攻击了，目前治疗只能治疗自己。
			log2 := boss.Attack(hero)

			response = log1 + "\n" + log2 + "\n"

		case "status":
			//太宣布事故Boss全局的状态血量
			response = fmt.Sprintf("状态：[%s] HP: %d/%d VS [%s] HP: %d/%d", hero.Name, hero.HP, hero.MaxHP, boss.Name, boss.HP, boss.MaxHP)

		case "say":
			if len(parts) < 2 {
				response = "格式错误，say <内容>\n"
				break
			}

			content := line[len(parts[0]):]

			content = strings.TrimSpace(content)

			msg := fmt.Sprintf("[%s]说 %s\n>", hero.Name, content)
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
			response = hero.ListInventory()

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
			ok, msg := hero.UnEquip(itemName)
			response = msg + "\n"
			if ok {
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
			return //踢走输掉的玩家
		}
		//最终战斗信息
		response += ">"
		conn.Write([]byte(response))

	}

}

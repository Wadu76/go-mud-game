package main

import (
	"bufio"
	"fmt"
	"mud-server/game"
	"os"
	"strings"
)

func main1() {

	//1初始化
	hero := game.NewPlayer("瓦度", 1, 100, 100)
	slime := game.NewMonster("史莱姆王", 50, 50, 20)

	fmt.Println("===== 欢迎来到 GO MUD 世界 ======")
	fmt.Println("请输入指令: attack, heal, status, exit")

	//2准备被读取器（从标准os.Stdin读取）
	reader := bufio.NewReader(os.Stdin)

	//3开始游戏循环
	for {
		fmt.Print("\n>") //打印提示符

		//读取用户输入直到按下回车
		input, _ := reader.ReadString('\n')

		//去掉输入前后的的换行符
		command := strings.TrimSpace(input)

		//4 处理指令
		switch command {
		case "attack":
			hero.Attack(slime)
			if slime.HP > 0 {
				slime.Attack(hero)
			}

		case "heal":
			//hero.Heal(hero)
			slime.Attack(hero)
		case "status":
			fmt.Printf(" 状态: [%s] HP: %d/%d\n", hero.Name, hero.HP, hero.MaxHP)
			fmt.Printf(" 敌人: [%s] HP: %d/%d\n", slime.Name, slime.HP, slime.MaxHP)
		case "exit":
			fmt.Println("游戏结束")
			return //结束游戏
		default:
			fmt.Println("无效指令, 请输入: attack, heal, status, exit")
		}

		//检测是否死亡
		if hero.HP <= 0 {
			fmt.Println("胜败乃兵家常事，请重新来过吧！💀")
			return //结束游戏
		} else if slime.HP <= 0 {
			fmt.Printf("史莱姆王已经死亡, 恭喜你！获取了胜利！，经验+%d\n", slime.Exp)
			return //结束游戏
		}
	}

}

package game

import (
	"encoding/json"
	"fmt"
	"os"
)

// 怪物的模板
// 用该结构体接json里的数据
type MonsterTemplate struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Exp    int    `json:"exp"`
	Attack int    `json:"attack"`
}

// 全局怪物模板库
// key是怪物id value为怪物模板数据
var MonsterTemplates map[string]MonsterTemplate

// 加载怪物数据
func LoadMonsterData(filePath string) error {
	//打开文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	//解析json
	var templates []MonsterTemplate
	//data是字节切片，包含要解析的json数据
	//templates是解析后存储的指向结构体的指针
	//json.Unmarshal是将json数据解析的函数
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return err
	}

	//转存到map中，后续可以按照id查找
	MonsterTemplates = make(map[string]MonsterTemplate)
	for _, t := range templates {
		MonsterTemplates[t.ID] = t
	}

	fmt.Printf("怪物数据加载成功! 共%d种\n", len(MonsterTemplates))
	return nil //返回nil表示没有错误

}

func NewMonsterFromID(id string) *Monster {
	//查找模板
	tpl, ok := MonsterTemplates[id]
	if !ok {
		fmt.Printf("找不到怪物ID %s, 使用默认的小史莱姆代替", id)
		return NewMonster("小史莱姆", 10, 10, 1, 1)
	}
	//找到后就创建具体的怪物对象
	return NewMonster(tpl.Name, tpl.HP, tpl.MaxHP, tpl.Exp, tpl.Attack)
}

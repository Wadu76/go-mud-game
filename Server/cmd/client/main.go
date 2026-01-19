package main

import (
	"encoding/json" //json解析库
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textinput" //输入框组件
	"github.com/charmbracelet/bubbles/viewport"  //滚动视窗组件
	tea "github.com/charmbracelet/bubbletea"     //核心引擎
	"github.com/charmbracelet/lipgloss"          //调色盘
)

// 定义样式
var (
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D9FF")).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00D9FF")).
			Padding(0, 1)

	styleInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// modle 数据结构体
type model struct {
	conn      net.Conn        //保存与服务器的连接
	viewport  viewport.Model  //聊天记录的滑动视图
	textInput textinput.Model //用户输入框
	err       error           //保存错误信息

	historyContent string //聊天记录
	ready          bool   //是否准备就绪 用于处理窗口初始化

	//玩家血量状态
	hp    int
	maxHp int

	//背包相关组件
	showInventory bool   //是否在显示背包
	inventory     []Item //背包里的东西

	//判断死亡组件，用于复活
	isDead bool //是否死亡
}

// 定义两个消息
type errMsg error     //错误消息
type serverMsg string //服务器发来的消息

// 定义跟服务器一样的结构体用来接收数据
type Item struct {
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Value      int    `json:"value"`
	IsEquipped bool   `json:"is_Equipped"`
}

// Init 初始化上述结构体中的内容
func initalModel() model {
	//初始化输入框 textinput ti
	ti := textinput.New()
	ti.Placeholder = "在此输入指令"
	ti.Focus()         //光标默认
	ti.CharLimit = 156 //限制输入长度
	ti.Width = 30      //设置输入框宽度 (稍微宽一点)

	//初始化视窗 viewport vp
	//此处先不初始化，等程序检测屏幕大小再初始化（update中）

	return model{
		textInput:      ti,
		historyContent: "正在连接瓦度世界...\n", //初始日志
		err:            nil,

		hp:    100,
		maxHp: 100,

		showInventory: false,
		inventory:     []Item{},
	}
}

// 连接服务器
func (m model) Init() tea.Cmd {
	//让光标聚焦到输入框，并且连接服务器
	return tea.Batch(textinput.Blink, connectToServer)
}

// Update，类比unity中的update
// 时刻更新时间，处理大部分事件（按下确认，收到服务器消息，报错等）
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	//窗口大小变化，刚打开程序的时候也有一次
	case tea.WindowSizeMsg:
		headerHeight := 2 //标题栏高度
		footerHeight := 3 //输入框高度 (留点空间给Help)
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			//这是第一次检测到窗口大小，即第一次打开程序，ready为false
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight     //从标题下面开始画
			m.viewport.SetContent(m.historyContent) //填入历史记录
			m.ready = true                          //第一次检测完毕
		} else {
			//窗口变换后就动态调整大小
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	//刚连接上服务器
	case net.Conn:
		m.conn = msg //保存连接
		//开始监听服务器有没有发消息
		return m, waitForServerMsg(m.conn)

	//收到了服务器的消息
	case serverMsg:
		//拿到初始消息
		fullText := string(msg)

		//拦截背包的数据
		if strings.Contains(fullText, "|CMD:INC:") {
			parts := strings.Split(fullText, "|CMD:INC:")
			if len(parts) > 1 {
				jsonStr := parts[1]
				var items []Item
				err := json.Unmarshal([]byte(jsonStr), &items)
				if err == nil {
					m.inventory = items
					m.showInventory = true
				}
			}
			return m, waitForServerMsg(m.conn)
		}

		//拦截死亡
		if strings.Contains(fullText, "|CMD:DIE") {
			m.isDead = true
			fullText = strings.ReplaceAll(fullText, "|CMD:DIE", "")
		}
		//拦截HP数据
		if strings.Contains(fullText, "|CMD:HP") {
			parts := strings.Split(fullText, "|CMD:HP")
			if len(parts) > 1 {
				fullText = parts[0] //保留文本
				//解析数值
				params := strings.Split(strings.TrimPrefix(parts[1], ":"), ":")
				if len(params) >= 3 {
					fmt.Sscanf(params[1], "%d", &m.hp)
					fmt.Sscanf(params[2], "%d", &m.maxHp)
				}
			}
		}

		// --- [修正] 暴力清洗文本，解决青色字乱飘问题 ---
		// 1. 兼容不同系统的换行
		fullText = strings.ReplaceAll(fullText, "\r\n", "\n")
		// 2. 暴力删除所有 ">" 提示符 (不论有没有空格)
		fullText = strings.ReplaceAll(fullText, "> ", "")
		fullText = strings.ReplaceAll(fullText, ">", "")
		// 3. 去除首尾空白，防止空行太多
		fullText = strings.TrimSpace(fullText)

		//如果洗完之后还有内容，才显示
		if fullText != "" {
			// [修正] 强制加换行 "\n"，保证每条消息独占一行！
			styledText := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")). //青色
				Render(fullText + "\n")           //简单粗暴加换行

			//把新消息加到历史记录中
			m.historyContent += styledText
			//把更新后的记录塞给视窗
			m.viewport.SetContent(m.historyContent)
			//自动滚到底部
			m.viewport.GotoBottom()
		}
		//听完一句继续监听下一句
		return m, waitForServerMsg(m.conn)

	//键盘输入，回车
	case tea.KeyMsg:
		//背包操作逻辑
		if m.showInventory || m.isDead {
			switch msg.String() {
			//esc q 或者再按次i关闭背包 允许回车复活
			case "esc", "q", "i", "enter":
				if m.isDead && msg.Type == tea.KeyEnter {
					m.isDead = false //复活
				}
				// 只有没死的时候才能关背包
				if !m.isDead {
					m.showInventory = false
				}
				return m, nil
			}
			return m, nil
		}

		switch msg.Type {
		//Ctrl+C退出
		case tea.KeyCtrlC:
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit

		//回车发送消息
		case tea.KeyEnter:
			//输入框的内容
			inputMsg := m.textInput.Value()
			//发送给服务器
			if m.conn != nil && inputMsg != "" {
				fmt.Fprintln(m.conn, inputMsg)

				// [修正] 解决灰色字不对齐
				// 不用 Align，直接在前面加 "\n" 强制换行，把灰色字顶到下一行去
				userMsg := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Render("\n> " + inputMsg + "\n")

				m.historyContent += userMsg
				m.viewport.SetContent(m.historyContent)
				m.viewport.GotoBottom()
			}

			//清空输入框
			m.textInput.Reset()
		}

	//发生错误
	case errMsg:
		m.err = msg
		return m, nil
	}

	//组件闪烁动画
	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)

}

// view 渲染 相当于unity的OnGUI 写了才能返回model
// 此处也是背包可视化主要逻辑所在之处
func (m model) View() string {
	//死亡弹窗，优先级最高
	if m.isDead {
		deadTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF0000")). // 红底白字
			Padding(1, 4).
			Render("💀  胜败乃兵家常事  💀")

		deadSubtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 Enter 复活并回城")

		// 居中显示
		return lipgloss.Place(m.viewport.Width, m.viewport.Height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, deadTitle, "\n", deadSubtitle),
		)
	}
	if !m.ready {
		return "\n 正在初始化界面..."
	}

	//渲染标题栏
	header := styleTitle.Render("Wadu MUD Client")

	//血条，按百分比来
	percent := float64(m.hp) / float64(m.maxHp)
	if percent < 0 {
		percent = 0
	} //防止血条为负

	if percent > 1 {
		percent = 1
	} //防止报表

	//血条宽度
	availableWidth := m.viewport.Width - 20
	maxBarWidth := 50
	//取二者较小
	barWidth := availableWidth
	if barWidth > maxBarWidth {
		barWidth = maxBarWidth
	}

	if barWidth < 10 {
		barWidth = 10
	} //最小宽度10

	//filled即当前血量
	filledCount := int(percent * float64(barWidth))

	//红色代表血
	filled := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(strings.Repeat("█", filledCount))
	//灰色代表空血
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render(strings.Repeat("░", barWidth-filledCount))

	hpBar := fmt.Sprintf("HP: [%s%s] %d/%d", filled, empty, m.hp, m.maxHp)

	//提示指令功能
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	helpText := " 指令: attack, heal, look, go <方向>, inventory, pick <物品>, equip <物品>"

	//渲染底部输入栏 横线 + 血条 + 输入框
	//使用lipgloss.JoinVertical 安全地垂直拼接，避免EXTRA string报错
	footer := lipgloss.JoinVertical(lipgloss.Left,
		styleInfo.Render(strings.Repeat("─", m.viewport.Width)), // 分割线
		hpBar,                      // 血条
		helpStyle.Render(helpText), // [新增] 帮助提示插在这里
		m.textInput.View(),         // 输入框
	)

	gameView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
	//如果没打开背包，直接返回正常界面
	if !m.showInventory {
		return gameView
	}

	//定义列  固定宽度，强制对齐
	//名字是青色
	colName := lipgloss.NewStyle().Width(14).Foreground(lipgloss.Color("#00FFFF"))
	//数值（力量）是红色
	colVal := lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Foreground(lipgloss.Color("#FF0000"))
	//描述是灰色
	colDesc := lipgloss.NewStyle().Width(30).Foreground(lipgloss.Color("#AAAAA"))

	//表头
	headerStr := lipgloss.JoinHorizontal(lipgloss.Top,
		colName.Render("名称"),
		colVal.Render("攻击"),
		"  ", //空格
		colDesc.Render("描述"),
	)

	//制作分割线
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#44444")).Render(strings.Repeat("-", 56))

	//制作列表内容
	var rows []string
	if len(m.inventory) == 0 {
		rows = append(rows, "   (背包空空如也...)")
	} else {
		for _, item := range m.inventory {
			//处理名字，装备了的加个 [E] equipped
			nameStr := item.Name
			if item.IsEquipped {
				nameStr = "[E]" + item.Name
			}

			//拼接
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				colName.Render(nameStr),                      //第一列名字
				colVal.Render(fmt.Sprintf("%d", item.Value)), //第二列数值
				"  ",                      //空格
				colDesc.Render(item.Desc), //第三列描述
			)
			rows = append(rows, row)
		}
	}

	//拼接所有内容 即整个表格
	tableBody := lipgloss.JoinVertical(lipgloss.Left, rows...)

	//最终样式
	inventoryWindow := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).             //双线边框
		BorderForeground(lipgloss.Color("#F1C40f")). //金色边框
		Padding(1, 2).                               //内边距
		Render(lipgloss.JoinVertical(lipgloss.Left,
			headerStr,
			divider,
			tableBody,
		))
	//绘制背包界面 (覆盖在上面)

	//返回背包界面，居中显示
	return lipgloss.JoinVertical(lipgloss.Center,
		header,
		"\n\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1C40F")).Render("=== 🎒 冒险者背包 ==="),
		inventoryWindow,
		"\n(按 ESC 关闭)",
	)

}

func connectToServer() tea.Msg {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		return errMsg(err)
	}
	return conn
}

func waitForServerMsg(conn net.Conn) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			return errMsg(err)
		}
		//把字节转换成字符串
		return serverMsg(string(buf[:n]))
	}
}

// 客户端的入口，与server中的main互不干扰
func main() {
	// AltScreen 模式：让程序像 Vim 一样占用整个屏幕，退出后自动恢复终端原状
	p := tea.NewProgram(initalModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}

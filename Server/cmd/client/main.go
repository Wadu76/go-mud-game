package main

import (
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

	
}

// 定义两个消息
type errMsg error     //错误消息
type serverMsg string //服务器发来的消息

// Init 初始化上述结构体中的内容
func initalModel() model {
	//初始化输入框 textinput ti
	ti := textinput.New()
	ti.Placeholder = "在此输入指令"
	ti.Focus()         //光标默认
	ti.CharLimit = 156 //限制输入长度
	ti.Width = 20      //设置输入框宽度

	//初始化视窗 viewport vp
	//vp := viewport.New(80, 20) //视窗大小，宽带80 高度20
	//vp.SetContent("正在连接瓦度世界...\n")
	//此处先不初始化，等程序检测屏幕大小再初始化（update中）
	//这样就可以避免输出过长导致无法输出完一整行

	return model{
		textInput: ti,
		//viewport:  vp,
		historyContent: "正在连接瓦度世界...\n", //初始日志
		err:            nil,
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
		footerHeight := 2 //输入框高度
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			//这是第一次检测到窗口大小，即第一次打开程序，ready为false
			//第一次检测到窗口大小的时候，初始化视窗
			//宽度 = 窗口宽度
			//高度 = 窗口高度 - 上下边距
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight     //从标题下面开始画
			m.viewport.SetContent(m.historyContent) //填入历史记录
			m.ready = true                          //第一次检测完毕，后续就不是了，因此设为true
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

		//新消息收录到历史记录中
		newText := string(msg)

		//服务器消息为青色
		//render 函数将文本渲染为带颜色的字符串
		styledText := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(newText)

		m.historyContent += styledText

		//把更新后的记录塞给视窗
		m.viewport.SetContent(m.historyContent)

		//自动滚到底部
		m.viewport.GotoBottom()

		//听完一句继续监听下一句
		return m, waitForServerMsg(m.conn)

	//键盘输入，回车
	case tea.KeyMsg:
		switch msg.Type {

		//Ctrl+C退出or ESC
		case tea.KeyCtrlC, tea.KeyEsc:
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
				fmt.Fprintf(m.conn, inputMsg+"\n")

				//自己发的也追加到历史记录
				//自己发的部分用灰色
				userlog := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("> " + inputMsg + "\n")
				m.historyContent += userlog
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
func (m model) View() string {
	if !m.ready {
		return "\n 正在初始化界面..."
	}

	//渲染标题栏
	header := styleTitle.Render("Wadu MUD Client")

	//渲染底部输入栏
	footer := fmt.Sprintf("%s\n%s",
		styleInfo.Render(strings.Repeat("-", m.viewport.Width)),
		m.textInput.View(),
	) //输入框的提示

	//头 + 视窗 + 尾
	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
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

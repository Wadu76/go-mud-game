using System;
using System.Collections; //引入协程需要的命名空间
using System.Net.Sockets;
using System.Text;
using System.Threading;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class NetworkManager : MonoBehaviour
{
    //暂时只写成 Wadu76，以后可以改为登录时获取
    private string playerName = "Wadu76";

    [Header("UI 基础组件")]
    public TMP_Text logText;        //显示日志的大屏幕
    public TMP_InputField inputField; //输入框
    public Button sendButton;       //发送按钮
    public ScrollRect chatScrollRect; //新增：用于控制滚动条自动到底

    [Header("战斗 UI")]
    public Slider hpSlider;      //血条

    [Header("死亡 UI")]
    public GameObject deathPanel; // 新增：死亡黑屏面板
    public Button reviveButton;   //复活按钮

    [Header("服务器配置")]
    public string serverIP = "127.0.0.1";
    public int serverPort = 8888;

    private TcpClient client;
    private NetworkStream stream;
    private Thread receiveThread;
    private bool isConnected = false;

    //线程安全队列
    private string messageBuffer = "";
    private object lockObj = new object();

    void Start()
    {
        ConnectToServer();

        //绑定发送按钮
        if (sendButton != null)
            sendButton.onClick.AddListener(OnSendButtonClicked);

        //新增：绑定复活按钮
        if (reviveButton != null)
            reviveButton.onClick.AddListener(OnReviveClicked);

        //确保一开始死亡面板是隐藏的
        if (deathPanel != null)
            deathPanel.SetActive(false);
    }

    //点击复活按钮的逻辑
    void OnReviveClicked()
    {
        SendMessageToServer("revive"); //发送复活指令给 Go
        if (deathPanel != null)
            deathPanel.SetActive(false); //隐藏黑屏
    }

    void ConnectToServer()
    {
        try
        {
            client = new TcpClient();
            client.Connect(serverIP, serverPort);
            stream = client.GetStream();
            isConnected = true;

            AddToLog("成功连接到瓦度世界！");

            receiveThread = new Thread(ReceiveData);
            receiveThread.IsBackground = true;
            receiveThread.Start();
        }
        catch (Exception e)
        {
            AddToLog("连接失败: " + e.Message);
        }
    }

    public void SendMessageToServer(string msg)
    {
        if (!isConnected) return;
        try
        {
            byte[] data = Encoding.UTF8.GetBytes(msg + "\n");
            stream.Write(data, 0, data.Length);
        }
        catch (Exception e)
        {
            AddToLog("发送失败: " + e.Message);
        }
    }

    void ReceiveData()
    {
        byte[] buffer = new byte[1024];
        while (isConnected)
        {
            try
            {
                if (stream.CanRead)
                {
                    int bytesRead = stream.Read(buffer, 0, buffer.Length);
                    if (bytesRead > 0)
                    {
                        string response = Encoding.UTF8.GetString(buffer, 0, bytesRead);
                        lock (lockObj)
                        {
                            messageBuffer += response;
                        }
                    }
                }
            }
            catch (Exception)
            {
                isConnected = false;
                break;
            }
        }
    }

    //主循环 (每帧调用) Upadte
    void Update()
    {
        lock (lockObj)
        {
            if (!string.IsNullOrEmpty(messageBuffer))
            {
                // 处理粘包和多条指令
                string rawMsg = messageBuffer;
                string[] parts = rawMsg.Split('|');

                // 第一部分通常是聊天内容
                if (!string.IsNullOrEmpty(parts[0]))
                {
                    //过滤掉单纯的提示符 "> "，不然太乱
                    if (parts[0] != "> " && parts[0] != ">")
                        logText.text += parts[0];

                    //只要有新文字，就自动滚动到底部
                    StartCoroutine(AutoScrollToBottom());
                }

                //后续部分是指令
                for (int i = 1; i < parts.Length; i++)
                {
                    string cmdPart = parts[i].Trim();

                    //血条指令
                    if (cmdPart.StartsWith("CMD:HP"))
                    {
                        string[] data = cmdPart.Split(':');
                        //格式: CMD:HP:Name:Cur:Max
                        if (data.Length >= 5)
                        {
                            string targetName = data[2];
                            if (int.TryParse(data[3], out int curHp) && int.TryParse(data[4], out int maxHp))
                            {
                                //只有名字是自己时，才更新左上角的血条
                                if (targetName == playerName)
                                {
                                    UpdateHPBar(curHp, maxHp);
                                }
                            }
                        }
                    }
                    //死亡指令
                    else if (cmdPart.StartsWith("CMD:DEAD"))
                    {
                        //收到死亡通知，显示黑屏面板
                        if (deathPanel != null)
                            deathPanel.SetActive(true);
                    }
                }

                messageBuffer = ""; //清空缓存
            }
        }

        if (Input.GetKeyDown(KeyCode.Return))
        {
            OnSendButtonClicked();
        }
    }

    void OnSendButtonClicked()
    {
        string txt = inputField.text;
        if (!string.IsNullOrEmpty(txt))
        {
            SendMessageToServer(txt);
            inputField.text = "";
            inputField.ActivateInputField();
        }
    }

    void AddToLog(string msg)
    {
        lock (lockObj)
        {
            messageBuffer += msg + "\n";
        }
    }

    //修复后的血条逻辑
    void UpdateHPBar(int current, int max)
    {
        if (hpSlider != null)
        {
            hpSlider.maxValue = max; //之前这里写错了，必须是 max
            hpSlider.value = current;
        }
    }

    //自动滚动到底部的协程
    IEnumerator AutoScrollToBottom()
    {
        //等待这一帧UI渲染结束
        yield return new WaitForEndOfFrame();

        if (chatScrollRect != null)
        {
            //强制把滚动条拉到最下面 (0 是底部，1 是顶部)
            chatScrollRect.verticalNormalizedPosition = 0f;
        }
    }

    void OnApplicationQuit()
    {
        isConnected = false;
        if (stream != null) stream.Close();
        if (client != null) client.Close();
    }
}
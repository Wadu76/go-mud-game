using System;
using System.Net.Sockets;
using System.Text;
using System.Threading;
using TMPro; //引用 TextMeshPro
using UnityEngine;
using UnityEngine.UI;

public class NetworkManager : MonoBehaviour
{
    [Header("UI组件")]
    public TMP_Text logText;        //显示日志的大屏幕
    public TMP_InputField inputField; //;输入框
    public Button sendButton;       //发送按钮

    [Header("血条组件")]
    public Slider hpSlider;      //用于控制血条增减

    [Header("服务器配置")]
    public string serverIP = "127.0.0.1"; //本地 IP
    public int serverPort = 8888;         //Go服务器监听的端口

    private TcpClient client;
    private NetworkStream stream;
    private Thread receiveThread;
    private bool isConnected = false;

    //线程安全队列（因为网络线程不能直接改 UI，要存起来让主线程改）
    private string messageBuffer = "";
    private object lockObj = new object();

    void Start()
    {
        ConnectToServer();

        //绑定按钮点击事件
        sendButton.onClick.AddListener(OnSendButtonClicked);
    }

    //连接服务器
    void ConnectToServer()
    {
        try
        {
            client = new TcpClient();
            client.Connect(serverIP, serverPort);
            stream = client.GetStream();
            isConnected = true;

            AddToLog("成功连接到瓦度世界！");

            //开启一个后台线程专门负责听服务器说话
            receiveThread = new Thread(ReceiveData);
            receiveThread.IsBackground = true;
            receiveThread.Start();
        }
        catch (Exception e)
        {
            AddToLog("连接失败: " + e.Message);
        }
    }

    //发送数据给 Go
    public void SendMessageToServer(string msg)
    {
        if (!isConnected) return;

        try
        {
            //Go那边是用 strings.TrimSpace 处理的，但最好还是加个换行符 \n 
            //确保 Go 知道这句话说完了
            byte[] data = Encoding.UTF8.GetBytes(msg + "\n");
            stream.Write(data, 0, data.Length);

            // 顺便把自己的话也显示在屏幕上
            // AddToLog("我: " + msg); 
        }
        catch (Exception e)
        {
            AddToLog("发送失败: " + e.Message);
        }
    }

    //接收数据 (在后台线程跑)
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
                        //把收到的消息存进缓存，等待主线程显示
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

    //主循环 (每帧调用，负责更新 UI)
    void Update()
    {
        //检查缓存里有没有新消息
        lock (lockObj)
        {
            if (!string.IsNullOrEmpty(messageBuffer))
            {
                string rawMsg = messageBuffer;

                //检查是否有|CMD
                if (rawMsg.Contains("|CMD:"))
                {
                    //按照|切割
                    string[] parts = rawMsg.Split('|');
                    if (parts.Length >= 2)
                    {
                        string chatPart = parts[0];
                        //Trim去除换行符号，以防万一
                        string cmdPart = parts[1].Trim();


                        if (cmdPart.StartsWith("CMD:HP"))
                        {
                            //再按冒号切分指令
                            string[] data = cmdPart.Split(':');
                            //ata[0]=CMD, data[1]=HP, data[2]=Name, data[3]=Cur, data[4]=MaxHP

                            if (data.Length >= 5)
                            {
                                if (int.TryParse(data[3], out int curHp) && int.TryParse(data[4], out int maxHp))
                                {
                                    UpdateHPBar(curHp, maxHp);
                                }
                            }
                        }


                        // logText.text += messageBuffer; //追加到大屏幕
                        logText.text += chatPart + "\n";
                        //messageBuffer = ""; //清空缓存
                    }
                    //自动滚动到底部 (如果用了 ScrollView 需要这行，现在不需要)
                }
                else
                {
                    logText.text += rawMsg;
                }

                messageBuffer = "";
            }

            //允许按回车发送
            if (Input.GetKeyDown(KeyCode.Return))
            {
                OnSendButtonClicked();
            }
        }
    }

    //按钮点击处理
    void OnSendButtonClicked()
    {
        string txt = inputField.text;
        if (!string.IsNullOrEmpty(txt))
        {
            SendMessageToServer(txt);
            inputField.text = ""; //清空输入框
            inputField.ActivateInputField(); //让光标回到输入框
        }
    }

    //辅助函数：把日志加到 buffer
    void AddToLog(string msg)
    {
        lock (lockObj)
        {
            messageBuffer += msg + "\n";
        }
    }

    //游戏关闭时断开连接
    void OnApplicationQuit()
    {
        isConnected = false;
        if (stream != null) stream.Close();
        if (client != null) client.Close();
    }


    //辅助函数：更新血条
    void UpdateHPBar(int current, int max)
    {
        if (hpSlider != null)
        {
            hpSlider.maxValue = max;
            hpSlider.value = current;
        }
    }
}

   
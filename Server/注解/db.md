*** DSN（Data Source Name）字符串： ***

- "root:123456"：数据库用户名和密码
- "@tcp(127.0.0.1:3306)"：数据库服务器地-址和端口（本地MySQL默认端口）
- "/mud_game"：要连接的数据库名称
- "?charset=utf8mb4"：指定字符集为utf8mb4，支持完整的UTF-8字符集
- "&parseTime=True"：将时间类型解析为time.Time类型
- "&loc=Local"：使用本地时区
# imagepool

一个基于云存储的图床服务，可以自动删除长时间无访问的图片。

分为服务端和客户端。

服务端负责接收访问请求，生成可短时间内访问图片的url并转发给请求者。拥有一个Redis数据库记录最近访问时间，定期删除那些长时间没有访问的图片。

客户端为一个命令行工具，上传图片到云存储，并通知服务端增加记录。

目前仅为我的 markdown 笔记提供图床服务。

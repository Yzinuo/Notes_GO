# 抓包

## 工具
- 把请求复制为Curl后，使用[curlconverter](curlconverter.com)就可以生成发送http请求的代码
- 序列化json数据，需要定义结构体。json数据如果庞大，自己定义结构体会很麻烦，可以使用[oktools](https://link.juejin.cn/)

# SOCKS5
## 历史
某些公司内部，十分注重数据安全，因此建立了**十分严密的防火墙**。但是因此呢，公司内部的内网络也很难访问外网络的信息。SOCKS5就是在这个墙中开了小口，让内网也可以访问外网。

## 原理：
![Alt text](/images/image.png)

- 1. 首先 Client和代理服务器建立链接，浏览器向代理服务器发送报文（协议版本号，鉴权方法的数量，健全的方法）
- 2. 代理服务器收到Client请求后，向目标服务器建立TCP链接
- 3. 代理服务器正常工作

第二步，代理服务器发送的报文：
// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	// VER 版本号，socks5的值为0x05
	// CMD 0x01表示CONNECT请求
	// RSV 保留字段，值为0x00
	// ATYP 目标地址类型，DST.ADDR的数据对应这个字段的类型。
	//   0x01表示IPv4地址，DST.ADDR为4个字节
	//   0x03表示域名，DST.ADDR是一个可变长度的域名
	// DST.ADDR 一个可变长度的值
	// DST.PORT 目标端口，固定2个字节

    
## 实现
```go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

)

func main(){
	server,err := net.Listen("tcp","127.0.0.1:1080")
	if err != nil {
		log.Fatal(err)
	}

	for{
		conn,err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed!!!!")
			continue
		}

		go process(conn)
	}
}

func process(conn net.Conn){
	defer conn.Close()
	reader := bufio.NewReader(conn)
	
	for{
		w,err := reader.ReadByte()
		if err != nil{
			break
		}

		_,err = conn.Write([]byte{w})
		if err!= nil{
			break
		}
	}
}
```

运行这个代码就需要run之后，新建终端输入**nc命令**建立链接。



package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)
const socks5Ver = 0x05
const cmdBind = 0x01
const atypeIPV4 = 0x01
const atypeHOST = 0x03
const atypeIPV6 = 0x04

func main(){
	server,err := net.Listen("tcp","127.0.0.1:1080")
	if err != nil {
		panic(err)
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
	
    err := auth(reader,conn)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("auth success!!!")

	err = connect(reader,conn)
	if err != nil {
		log.Fatal(err)
	}
}

// 1. 客户端向代理服务器发送认证请求
func auth(reader *bufio.Reader,conn net.Conn)error{
	ver,err := reader.ReadByte()
	if err != nil {
		return err
	}
	if ver != socks5Ver {
		return errors.New("unsupport ver")
	}

	methodsize,err := reader.ReadByte()
	if err !=nil{
		return err
	}

	methods := make([]byte,methodsize)
	_,err = io.ReadFull(reader,methods)
	if err!=nil{
		return err
	}

	log.Println("version: ",ver,"method:",methods)
	_,err = conn.Write([]byte{socks5Ver,0x00})
	if err != nil {
		return errors.New("write failed")
	}

	return nil
}

// 2. 代理服务器向服务器发送链接请求
// 3. 代理服务器向服务器发送响应,从服务器读取数据，发送给客户端

func connect(reader *bufio.Reader,conn net.Conn) (err error){
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
	cache := make([]byte,4)
	_,err = io.ReadFull(reader,cache)
	if err != nil {
		return errors.New("read failed function io.readfull")
	}

	// 判断每一个协议的值
	if cache[0]!= socks5Ver || cache[1]!= cmdBind {
		return errors.New("unsupport ver or cmd")
	}

	atyp := cache[3]
	addr := ""
	
	switch atyp {
	case atypeIPV4:
	      _,err := io.ReadFull(reader,cache)
	      if err!= nil {
	      	return errors.New("read failed function io.readfull at astypeIPV4")
	      }
		  addr = fmt.Sprintf("%d.%d.%d.%d",cache[0],cache[1],cache[2],cache[3])

	case atypeIPV6:
		return errors.New("unsupport atypeIPV6")

	case atypeHOST:
	      hostsize,err := reader.ReadByte()
	      if err!= nil {
			return errors.New("read failed function readbyte at astypeHOST")	
		  }

		  newcache := make([]byte,hostsize)
		  _,err = io.ReadFull(reader,newcache)
		  if err!= nil {
			return errors.New("read failed function io.readfull at astypeHOST")
		  }
		  addr = string(newcache)
	default:
		return errors.New("invaild atyp")
	}
	
	_,err = io.ReadFull(reader,cache[:2])
	if err!= nil {
		return errors.New("read failed function io.readfull at port")
	}
	port := binary.BigEndian.Uint16(cache[:2])

	log.Printf("addr: %s,port: %d",addr,port)

	// 建立和服务器的链接
	dest,err := net.Dial("tcp",fmt.Sprintf("%s:%d",addr,port))
	
	if err!= nil {
		return errors.New("dial failed")
	}
	defer dest.Close()
	log.Println("dial", addr, port)

	
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err!= nil {
		return errors.New("write failed")
	}

	// 实现数据交换！
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(){
		_,_ = io.Copy(dest,reader)
		cancel()
	}()

	go func(){
		_,_ = io.Copy(conn,dest)
		cancel()
	}()
	
	// 任意一个协程失败，就把两个goroutine都结束，就结束
	<-ctx.Done()
	return nil
}


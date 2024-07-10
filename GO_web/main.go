package main

import "net/http"

func main(){
	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
			w.Write([]byte("fuck world!!!"))	
	}) // "/"根地址,代表对所有请求响应
	
	// 启动HTTP服务器，监听8080端口。当有请求到达时，服务器会根据请求的路径调用相应的处理函数。
	http.ListenAndServe("localhost:8080",nil)//nil代表使用DefaultSeverMux
	
}
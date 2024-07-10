当有http请求时,**handler就会创建一个goroutine去处理请求**
![Alt text](image.png)

GO中默认的handler就是http.DefalutseverMux,作用类似路由(它的作用是将传入的 HTTP 请求根据请求的 URL 路径分发到相应的处理函数)
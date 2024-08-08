package log

import (
	"io"
	stllog "log"
	"net/http"
	"os"
)

// 自定义log的需要
var log *stllog.Logger

// 绑定方法不能是内置类型，必需自己声明
type fileLog string

func (fs fileLog) Write(data []byte) (int,error){
	f,err := os.OpenFile(string(fs),os.O_CREATE | os.O_WRONLY | os.O_APPEND,0660)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	
	return f.Write(data)
}
 

func Run(destination string){
	// 第一个参数是写入目的地，要求是io.Writer接口,第二个事前缀
	log = stllog.New(fileLog(destination),"go",stllog.LstdFlags)
}

func RegisterHandler(){
	http.HandleFunc("/log",func(w http.ResponseWriter, r *http.Request){
		switch r.Method{
		case http.MethodPost:
			msg,err := io.ReadAll(r.Body)
			if err != nil || len(msg) == 0{
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}) 
}

func write (s string){
	//使用自己定义的log的printf方法，把数据写入目标文件
	log.Printf("%v\n",s)
}
 





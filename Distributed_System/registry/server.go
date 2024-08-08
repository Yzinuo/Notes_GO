package registry

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/Services"

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

func (r *registry) add(n Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, n)
	r.mutex.Unlock()

	//注册的时候是添加依赖的最好时机，把依赖的服务请求过来
	err := r.sendRequiredServices(n)

	// 添加依赖变化逻辑
	r.Notify(patch{
		Added: []patchEntry{{
			Name : n.ServiceName,
			URL  : n.ServiceURL, 
		},
		},
	})

	return err
}

func (r registry) Notify(pat patch){
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _,reg := range r.registrations{
		go func(reg Registration){
			for _,reqService := range reg.RequiredServices{
				p := patch{Added : []patchEntry{}, Removed: []patchEntry{}}
				Sendupdate := false
				for _,added := range pat.Added{
					if added.Name == reqService{
						Sendupdate = true
						p.Added = append(p.Added, added)
					}
				}
				for _,removed := range pat.Removed{
					if removed.Name == reqService{
						Sendupdate = true
						p.Removed = append(p.Removed, removed)
					}
				}
				if Sendupdate{
					err := r.SendPatch(p,reg.ServiceUpdateURL)
					if err != nil{
						log.Println(err)
						return
					}
				}
			}
			
		}(reg)
	}
}

func (r registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.SendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}

	return nil
}

func (r registry) SendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	return nil
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{}

func (s *RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request Received!")

	switch r.Method {
	case http.MethodPost:
		var re Registration
		err := json.NewDecoder(r.Body).Decode(&re)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = reg.add(re)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// 听心跳：检测服务状态。这里我们用通过向每个service发送post请求的方法验证他们是否正常工作的方法。
func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _,reg := range r.registrations{
			wg.Add(1)
			go func (reg Registration){
				defer wg.Done()
				success := true
				
				// 心跳检查默认三次机会
				for attemps := 0; attemps <3;attemps++ {
					res ,err := http.Get(reg.heartbeatURL)
					if err != nil{
						log.Println(err)
					}else if res.StatusCode == http.StatusOK {
						if(!success){
							r.add(reg)
						}
						break
					}
					log.Println("HeartBeat check failed")
					if(success){
						success = false
						r.remove(reg)
					}
					time.Sleep(1*time.Second)
				}
				
			}(reg)
		}
		wg.Wait()
		time.Sleep(freq)
	}
}

// 保证仅运行一次
var once sync.Once

func CheckHeartBeat(){
	once.Do(func(){
		go  reg.heartbeat(3 * time.Second)
	})
}
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

func RegisterService(r Registration) error {
	// patch发送给UpdateURL add remove数据，编写这个UpdateURL的handler
	serviceUpdateUrl,err := url.Parse(r.ServiceUpdateURL)
	if err != nil{
		return err
	}
	http.Handle(serviceUpdateUrl.Path,&serviceUpdateHandler{})

	heartbeatUrl,err := url.Parse(r.heartbeatURL)
	if err != nil{
		return err
	}

	http.HandleFunc(heartbeatUrl.Path,func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusOK)
	})
	
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(&r)
	if err != nil {
		log.Println("fail to encode!")
		return err
	}

	res, err := http.Post(ServicesURL, "application/json", buf)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to registry service.")
	}

	log.Printf("Adding Service :%v with %v!!!", r.ServiceName, r.ServiceURL)
	return nil
}

type serviceUpdateHandler struct {}

func (serv *serviceUpdateHandler)ServeHTTP(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p patch
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil{
		log.Println("Decode unsuccess!")
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	prov.Update(p)
}
// 被依赖的Service需要通过发送请求获得，此时我们也需要一个地方存储被依赖的服务。
type Providers struct {
	service map[ServiceName][]string
	mutex   *sync.RWMutex
}

func (p Providers) Update(pat patch) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, patchEntry := range pat.Added {
		if _, ok := p.service[patchEntry.Name]; !ok {
			// map为新key创造新的value
			p.service[patchEntry.Name] = make([]string, 0)
		}
		p.service[patchEntry.Name] = append(p.service[patchEntry.Name], patchEntry.URL)
	}

	for _, patchEntry := range pat.Removed {
		if providersURL, ok := p.service[patchEntry.Name]; ok {
			for i := range providersURL {
				if providersURL[i] == patchEntry.URL {
					p.service[patchEntry.Name] = append(p.service[patchEntry.Name][:i], p.service[patchEntry.Name][i+1:]...)
				}
			}
		}
	}

}

func (p Providers) get(name ServiceName) (string, error) {
	providers, ok := p.service[name]
	if !ok {
		return " ", fmt.Errorf("No providers avaliable for service %v", name)
	}

	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

var prov = Providers{
	// 一个服务可能由多个服务点提供，有多个URL
	service: make(map[ServiceName][]string),
	mutex:   new(sync.RWMutex),
}

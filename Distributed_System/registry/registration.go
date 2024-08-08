package registry

// 每一个web service应该记录的信息
type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	//服务发现的必要,记录依赖的服务
	RequiredServices []ServiceName
	//通过这个URL告诉当前service，依赖的service有无。
	ServiceUpdateURL string
	heartbeatURL	string
}

type ServiceName string

// 记录已有的web service
const (
	LogService = ServiceName("LogServive")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}

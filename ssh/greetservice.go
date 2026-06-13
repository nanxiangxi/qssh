package ssh

type GreetService struct {
	version string
}

// NewGreetService 创建服务实例
func NewGreetService(version string) *GreetService {
	return &GreetService{version: version}
}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}

// GetVersion 获取软件版本
func (g *GreetService) GetVersion() string {
	return g.version
}

// GetAppName 获取应用名称
func (g *GreetService) GetAppName() string {
	return "启SSH"
}

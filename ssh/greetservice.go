package ssh

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}

// GetVersion 获取软件版本
func (g *GreetService) GetVersion() string {
	return "0.2.0"
}

// GetAppName 获取应用名称
func (g *GreetService) GetAppName() string {
	return "启SSH"
}

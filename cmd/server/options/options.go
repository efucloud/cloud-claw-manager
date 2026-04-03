package options

type ServerRunOptions struct {
	Config          string
	KubeConfig      string
	LeaderNamespace string
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{}
	return &s
}

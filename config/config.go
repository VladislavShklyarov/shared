package config

type JaegerConfig struct {
	ServiceName string

	SamplerType  string
	SamplerParam float64

	AgentHost string
	AgentPort string

	LogSpans bool
}

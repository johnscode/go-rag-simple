package main

type config struct {
	Port         int    `env:"PORT" envDefault:"4000" json:"port,omitempty"`
	LogDir       string `env:"LOGDIR,expand" envDefault:"${HOME}/logs" json:"logDir,omitempty"`
	WeaviateHost string `env:"WEAVIATEHOST" envDefault:"localhost" json:"weaviateHost,omitempty"`
	WeaviatePort string `env:"WEAVIATEPORT" envDefault:"9035" json:"weaviatePort,omitempty"`
}

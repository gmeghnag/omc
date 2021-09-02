package models

type Context struct {
	Id      string `json:"id"`
	Path    string `json:"path"`
	Current string `json:"current"`
	Project string `json:"project"`
}

type Config struct {
	Id       string    `json:"id"`
	Contexts []Context `json:"contexts"` // <== contexts è la chiave json per accedere a []Context
}

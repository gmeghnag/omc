package types

import "k8s.io/client-go/kubernetes"

type Context struct {
	Id      string `json:"id"`
	Path    string `json:"path"`
	Current string `json:"current"`
	Project string `json:"project"`
}

type Config struct {
	Id       string    `json:"id,omitempty"`
	Contexts []Context `json:"contexts,omitempty"`
}

type DescribeClient struct {
	Namespace string
	kubernetes.Interface
}

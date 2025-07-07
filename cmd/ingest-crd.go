package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cliprint "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var IngestCRDS = &cobra.Command{
	Use:     "collect-crd",
	Aliases: []string{"ingest-crd", "ingest-crds", "collect-crds"},
	Short:   "Collect CRDs from your running cluster to ~/.omc/customresourcedefinitions/* .",
	Run: func(cmd *cobra.Command, args []string) {

		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		kubeconfigPath := os.Getenv("KUBECONFIG")

		if kubeconfigPath == "" {
			kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
		}
		outputDir := filepath.Join(homeDir, ".omc", "customresourcedefinitions")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			panic(err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		crdList, err := dynamicClient.Resource(getCRDGroupVersionResource()).Namespace("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			panic(err)
		}
		for _, crd := range crdList.Items {
			saveCRDToFile(crd, outputDir)
		}
	},
}

func getCRDGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
}

func saveCRDToFile(crd unstructured.Unstructured, outputDir string) {
	name := crd.GetName()
	filename := filepath.Join(outputDir, strings.ToLower(name)+".yaml")
	newFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	printer := cliprint.YAMLPrinter{}
	err = printer.PrintObj(&crd, newFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("Saved:", filename)
}

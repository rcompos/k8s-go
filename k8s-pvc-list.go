package main

import (
	"fmt"
	"os"
	"log"
	"path/filepath"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func listPVCs(ns string, c *kubernetes.Clientset, p *v1.PersistentVolumeClaimList) {
	if len(p.Items) == 0 {
		log.Println("No PVCs!")
		return
	}
    template := "%-32s%-8s%-8s\n"
    fmt.Println("--- PVCs ----")
    fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
    //var cap resource.Quantity
    for _, pvc := range p.Items {
        quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
        //cap.Add(quant)
        fmt.Printf(template, pvc.Name, string(pvc.Status.Phase), quant.String())
    }

    fmt.Println("-----------------------------")
    //fmt.Printf("Total capacity claimed: %s\n", cap.String())
    fmt.Println("-----------------------------")
}

//func GetPVCs(ns string, c *kubernetes.Clientset, sf string) *v1.PersistentVolumeClaimList {
func GetPVCs(ns string, c *kubernetes.Clientset) *v1.PersistentVolumeClaimList {
    //pvcs, err := c.CoreV1().PersistentVolumeClaims(string(ns)).List(metav1.ListOptions{
        //FieldSelector: "status.phase!=Running",
        //FieldSelector: sf,
    //})
    pvcs, err := c.CoreV1().PersistentVolumeClaims(string(ns)).List(metav1.ListOptions{})
    if err != nil {
        log.Fatalln("failed to get PVCs:", err)
    }
    //fmt.Println("PVCs: ", pvcs)
    return pvcs
}

func main() {

	var ns string 
	ns = "default"
    // create the clientset

    // Bootstrap k8s configuration from local Kubernetes config file
    kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        log.Fatal(err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    pvcs := GetPVCs(ns, clientset) 
    if err != nil {
        log.Fatal(err)
    }
	listPVCs(ns, clientset, pvcs)
	fmt.Println()

}

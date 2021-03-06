package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	_ "regexp"
	"strings"
	//"reflect"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//
// List and optionally delete Kubernetes pods 
// of specified non-running status

func main() {
	var ns string
	var nsall, Delete, Force bool
	flag.StringVar(&ns, "n", "default", "Define namespace")
	flag.BoolVar(&nsall, "a", false, "All namespaces")
	flag.BoolVar(&Delete, "d", false, "Delete pods")
	flag.BoolVar(&Force, "f", false, "Force without confirmation")
	flag.Parse()

	var selectField string
	selectField = "status.phase!=Running"

	// Bootstrap k8s configuration from local Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	if nsall == true {
		// All namespaces
		namespaces := GetNamespaces(clientset)
		for _, thisName := range namespaces.Items {
			HandlePrompt(thisName.Name, clientset, selectField, Delete, Force)
		}
	} else {
		// Single namespace
		HandlePrompt(ns, clientset, selectField, Delete, Force)
	}

} // func main

func HandlePrompt(ns string, c *kubernetes.Clientset, sf string, del bool, force bool) *v1.PodList {
	PodsAll := GetPods(ns, c, sf)
	PodsNum := len(PodsAll.Items)
	if PodsNum > 0 {
		if del != true {
			ListPods(ns, c, PodsAll)
		} else {
			if force != true {
				ListPods(ns, c, PodsAll)
				Prompter()
			}
			DeletePods(ns, c, PodsAll)
		}
	}
	return PodsAll
}

func GetNamespaces(c *kubernetes.Clientset) *v1.NamespaceList {
	namespaces, err := c.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get namespaces:", err)
	}
	return namespaces
}

func GetPods(ns string, c *kubernetes.Clientset, sf string) *v1.PodList {
	pods, err := c.CoreV1().Pods(string(ns)).List(metav1.ListOptions{
		//FieldSelector: "status.phase!=Running",
		FieldSelector: sf,
	})
	if err != nil {
		log.Fatalln("failed to get pods:", err)
	}
	return pods
}

func ListPods(ns string, c *kubernetes.Clientset, pods *v1.PodList) {
	for _, pod := range pods.Items {
		pname := pod.Name
		outc, err := json.Marshal(pname)
		if err != nil {
			panic(err)
		}
		TargetPod := trimQuote(string(outc))
		fmt.Printf("To be deleted: %s %s\n", ns, TargetPod)
	}
}

func DeletePods(ns string, c *kubernetes.Clientset, pods *v1.PodList) {
	for _, pod := range pods.Items {
		pname := pod.Name
		outc, err := json.Marshal(pname)
		if err != nil {
			panic(err)
		}
		TargetPod := trimQuote(string(outc))
		errDelete := c.CoreV1().Pods(ns).Delete(TargetPod, &metav1.DeleteOptions{})
		if errDelete != nil {
			log.Printf("Error deleting pod %s, %s", TargetPod, errDelete)
		}
		fmt.Println("Deleted! ", ns, TargetPod)
	}
}

func trimQuote(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func Prompter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Type y to continue: ")
	text, _ := reader.ReadString('\n')
	answer := strings.TrimRight(text, "\n")
	//fmt.Printf("answer: %s \n", answer)
	if answer == "y" || answer == "Y" {
		return
	} else {
		//Prompter() //For recursive prompting
		log.Fatal("Exiting without action.")
	}
}


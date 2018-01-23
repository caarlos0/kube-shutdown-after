package main

import (
	"log"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var replicas int32

func init() {
	log.SetPrefix("kube-shutdown-after: ")
}

func main() {
	// creates the in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/carlos/.kube/config")
	if err != nil {
		log.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		deploys, err := clientset.AppsV1beta2().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		now := time.Now().Local()

		for _, deploy := range deploys.Items {
			value, ok := deploy.Annotations["carlosbecker.com/shutdown-after"]
			if !ok {
				log.Println("deployment is not annotated, ignoring:", deploy.Name)
				continue
			}
			log.Printf("deployment %s is annotated with %s", deploy.GetName(), value)
			t, err := time.Parse("15:04", value)
			if err != nil {
				log.Printf("%s is not in 15:04 format", value)
				continue
			}
			if t.Hour() >= now.Hour() && t.Minute() >= now.Minute() {
				log.Printf("its time, scaling down %s", deploy.GetName())
				deploy.Spec.Replicas = &replicas
				_, err := clientset.
					AppsV1beta2().
					Deployments(deploy.GetNamespace()).
					Update(&deploy)
				if err != nil {
					log.Println("failed to scale down:", err)
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

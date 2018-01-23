package main

import (
	"flag"
	"log"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	zeroReplicas int32

	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
)

func init() {
	log.SetPrefix("kube-shutdown-after: ")
}

func getConfig(cfg string) (*rest.Config, error) {
	if *kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", cfg)
}

func main() {
	flag.Parse()

	config, err := getConfig(*kubeconfig)
	if err != nil {
		log.Fatalln("failed to get config:", err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		deploys, err := clientset.AppsV1beta2().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get deployments:", err)
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
				log.Printf("failed to parse `%s`: not in `15:04` format", value)
				continue
			}
			if t.Hour() >= now.Hour() && t.Minute() >= now.Minute() && *deploy.Spec.Replicas > 0 {
				log.Printf("scaling down %s", deploy.GetName())
				deploy.Spec.Replicas = &zeroReplicas
				_, err := clientset.
					AppsV1beta2().
					Deployments(deploy.GetNamespace()).
					Update(&deploy)
				if err != nil {
					log.Printf("failed to scale %s down: %s", deploy.GetName(), err)
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

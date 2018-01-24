package main

import (
	"flag"
	"log"
	"time"

	"k8s.io/api/apps/v1beta1"
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
		deploys, err := clientset.AppsV1beta1().Deployments(apiv1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalln("failed to get deployments:", err)
		}
		for _, deploy := range deploys.Items {
			if !shouldScale(deploy) {
				continue
			}
			if err := scaleDown(clientset, deploy); err != nil {
				log.Printf("failed to scale %s down: %s", deploy.GetName(), err)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func shouldScale(deploy v1beta1.Deployment) bool {
	now := time.Now().Local()
	value, ok := deploy.Annotations["shutdown-after"]
	if !ok {
		log.Println("deployment is not annotated, ignoring:", deploy.Name)
		return false
	}
	log.Printf("deployment %s is annotated with %s", deploy.GetName(), value)
	t, err := time.Parse("15:04", value)
	if err != nil {
		log.Printf("failed to parse `%s`: not in `15:04` format", value)
		return false
	}
	return t.Hour() >= now.Hour() && t.Minute() >= now.Minute() && *deploy.Spec.Replicas > 0
}

func scaleDown(clientset *kubernetes.Clientset, deploy v1beta1.Deployment) error {
	log.Printf("scaling down %s", deploy.GetName())
	deploy.Spec.Replicas = &zeroReplicas
	_, err := clientset.AppsV1beta1().Deployments(deploy.GetNamespace()).
		Update(&deploy)
	return err
}

func getConfig(cfg string) (*rest.Config, error) {
	if *kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", cfg)
}

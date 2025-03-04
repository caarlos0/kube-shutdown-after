package main

import (
	"context"
	"flag"
	"log"
	"time"

	v1 "k8s.io/api/apps/v1"
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
	debug      = flag.Bool("debug", false, "enable debug logs")
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
		ctx := context.Background()
		deploys, err := clientset.AppsV1().Deployments(apiv1.NamespaceAll).List(ctx, metav1.ListOptions{})
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
		time.Sleep(10 * time.Second)
	}
}

func shouldScale(deploy v1.Deployment) bool {
	value, ok := deploy.Annotations["shutdown-after"]
	if !ok {
		if *debug {
			log.Println("deployment is not annotated, ignoring:", deploy.Name)
		}
		return false
	}
	if *debug {
		log.Printf("deployment %s is annotated with %s", deploy.GetName(), value)
	}
	return *deploy.Spec.Replicas > 0 && isItTimeToScaleDown(value)
}

const format = "15:04 -07"

func isItTimeToScaleDown(value string) bool {
	t, err := time.Parse(format, value)
	if err != nil {
		log.Printf("failed to parse `%s`: not in `%s` format", value, format)
		return false
	}
	now := time.Now().In(t.Location())
	if *debug {
		log.Printf("t=%s, now=%s", t.Format(format), now.Format(format))
	}
	return now.Hour() == t.Hour() && now.Minute() == t.Minute()
}

func scaleDown(clientset *kubernetes.Clientset, deploy v1.Deployment) error {
	log.Printf("scaling down %s", deploy.GetName())
	deploy.Spec.Replicas = &zeroReplicas
	_, err := clientset.AppsV1().Deployments(deploy.GetNamespace()).
		Update(context.Background(), &deploy, metav1.UpdateOptions{})
	return err
}

func getConfig(cfg string) (*rest.Config, error) {
	if *kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", cfg)
}

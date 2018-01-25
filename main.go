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
		time.Sleep(30 * time.Second)
	}
}

func shouldScale(deploy v1beta1.Deployment) bool {
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

func isItTimeToScaleDown(value string) bool {
	t, err := time.Parse("15:04 MST", value)
	if err != nil {
		log.Printf("failed to parse `%s`: not in `15:04 MST` format", value)
		return false
	}
	// truncate to have only HH:mm, on the same TZ as the annotated time
	now := time.Now().In(t.Location()).Truncate(time.Minute)
	// t will initially be something like:
	// 0000-01-01 19:50:00 -0200 GMT-2
	// so add now.year, now.month and now.day, but remove 1 month and
	// 1 day (because it is 0000-01-01)
	t = t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1).Truncate(time.Minute)
	if *debug {
		log.Printf("t is %s and now is %s", t, now)
	}
	// shut down only if the time is the same (HH:mm), so if someone
	// is working late and scale the deployment back up, we will not scale
	// it down again and again and again
	return now.Equal(t)
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

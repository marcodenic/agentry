package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/marcodenic/agentry/internal/taskqueue"
	"github.com/marcodenic/agentry/pkg/autoscale"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	natsURL := flag.String("nats", natsURL(), "NATS server URL")
	subject := flag.String("subject", "agentry.tasks", "task subject")
	namespace := flag.String("namespace", "default", "k8s namespace")
	deployment := flag.String("deployment", "agentry-worker", "deployment name")
	min := flag.Int("min", 1, "minimum replicas")
	max := flag.Int("max", 5, "maximum replicas")
	factor := flag.Int("factor", 10, "messages per replica")
	poll := flag.Duration("poll", 5*time.Second, "poll interval")
	flag.Parse()

	q, err := taskqueue.NewQueue(*natsURL, *subject)
	if err != nil {
		panic(err)
	}
	defer q.Close()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	as := autoscale.New(q, client, autoscale.Options{
		Namespace:    *namespace,
		Deployment:   *deployment,
		MinReplicas:  *min,
		MaxReplicas:  *max,
		ScaleFactor:  *factor,
		PollInterval: *poll,
	})
	if err := as.Run(context.Background()); err != nil {
		panic(err)
	}
}

func natsURL() string {
	if u := os.Getenv("NATS_URL"); u != "" {
		return u
	}
	return "nats://localhost:4222"
}

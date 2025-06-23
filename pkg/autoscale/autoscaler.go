package autoscale

import (
    "context"
    "fmt"
    "time"

    "k8s.io/apimachinery/pkg/types"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

// Queue represents a task queue exposing lag information.
type Queue interface {
    // Lag returns the approximate number of pending tasks.
    Lag(ctx context.Context) (int, error)
}

// Options configures the Autoscaler.
type Options struct {
    Namespace    string
    Deployment   string
    MinReplicas  int
    MaxReplicas  int
    ScaleFactor  int           // messages per replica
    PollInterval time.Duration
}

// Autoscaler monitors queue lag and scales a Deployment accordingly.
type Autoscaler struct {
    q      Queue
    client kubernetes.Interface
    opts   Options
}

// New creates a new Autoscaler.
func New(q Queue, client kubernetes.Interface, opts Options) *Autoscaler {
    if opts.PollInterval == 0 {
        opts.PollInterval = 5 * time.Second
    }
    if opts.ScaleFactor <= 0 {
        opts.ScaleFactor = 10
    }
    if opts.MinReplicas <= 0 {
        opts.MinReplicas = 1
    }
    if opts.MaxReplicas < opts.MinReplicas {
        opts.MaxReplicas = opts.MinReplicas
    }
    return &Autoscaler{q: q, client: client, opts: opts}
}

// Run starts the autoscaling loop until ctx is canceled.
func (a *Autoscaler) Run(ctx context.Context) error {
    ticker := time.NewTicker(a.opts.PollInterval)
    defer ticker.Stop()
    for {
        if err := a.scaleOnce(ctx); err != nil {
            // log but continue
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
        }
    }
}

func (a *Autoscaler) scaleOnce(ctx context.Context) error {
    lag, err := a.q.Lag(ctx)
    if err != nil {
        return err
    }
    desired := lag/a.opts.ScaleFactor + 1
    if desired < a.opts.MinReplicas {
        desired = a.opts.MinReplicas
    }
    if desired > a.opts.MaxReplicas {
        desired = a.opts.MaxReplicas
    }

    dep, err := a.client.AppsV1().Deployments(a.opts.Namespace).Get(ctx, a.opts.Deployment, metav1.GetOptions{})
    if err != nil {
        return err
    }
    current := int32(1)
    if dep.Spec.Replicas != nil {
        current = *dep.Spec.Replicas
    }
    if int32(desired) == current {
        return nil
    }
    patch := []byte(fmt.Sprintf(`{"spec":{"replicas":%d}}`, desired))
    _, err = a.client.AppsV1().Deployments(a.opts.Namespace).Patch(ctx, a.opts.Deployment, types.MergePatchType, patch, metav1.PatchOptions{})
    return err
}


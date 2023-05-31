package leader

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Callback func(cb context.Context)

const defaultLeaseDuration = 45 * time.Second
const defaultRenewDeadline = 30 * time.Second
const defaultRetryPeriod = 2 * time.Second

const developmentLeaseDuration = 45 * time.Hour
const developmentRenewDeadline = 30 * time.Hour

func RunOrDie(ctx context.Context, namespace, name string, client kubernetes.Interface, cb Callback) {
	if namespace == "" {
		namespace = "kube-system"
	}

	err := run(ctx, namespace, name, client, cb)
	if err != nil {
		logrus.Fatalf("Failed to start leader election for %s", name)
	}
	panic("Failed to start leader election for " + name)
}

func run(ctx context.Context, namespace, name string, client kubernetes.Interface, cb Callback) error {
	id, err := os.Hostname()
	if err != nil {
		return err
	}

	rl, err := resourcelock.New(resourcelock.ConfigMapsLeasesResourceLock,
		namespace,
		name,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		logrus.Fatalf("error creating leader lock for %s: %v", name, err)
	}

	leaseDuration := defaultLeaseDuration
	renewDeadline := defaultRenewDeadline
	retryPeriod := defaultRetryPeriod
	if d := os.Getenv("CATTLE_DEV_MODE"); d != "" {
		leaseDuration = developmentLeaseDuration
		renewDeadline = developmentRenewDeadline
	}
	if d := os.Getenv("CATTLE_ELECTION_LEASE_DURATION"); d != "" {
		leaseDuration, err = time.ParseDuration(d)
		if err != nil {
			return errors.Wrapf(err, "CATTLE_ELECTION_LEASE_DURATION is not a valid duration")
		}
	}
	if d := os.Getenv("CATTLE_ELECTION_RENEW_DEADLINE"); d != "" {
		renewDeadline, err = time.ParseDuration(d)
		if err != nil {
			return errors.Wrapf(err, "CATTLE_ELECTION_RENEW_DEADLINE is not a valid duration")
		}
	}
	if d := os.Getenv("CATTLE_ELECTION_RETRY_PERIOD"); d != "" {
		retryPeriod, err = time.ParseDuration(d)
		if err != nil {
			return errors.Wrapf(err, "CATTLE_ELECTION_RETRY_PERIOD is not a valid duration")
		}
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				go cb(ctx)
			},
			OnStoppedLeading: func() {
				logrus.Fatalf("leaderelection lost for %s", name)
			},
		},
		ReleaseOnCancel: true,
	})
	panic("unreachable")
}

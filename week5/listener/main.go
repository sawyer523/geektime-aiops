package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
)

func main() {
	home := homedir.HomeDir()
	if home != "" {
		home = filepath.Join(home, ".kube", "config")
	}

	kubeconfig := flag.String("kubeconfig", home, "absolute path to the kubeconfig file")
	flag.Parse()

	var (
		config *rest.Config
		err    error
	)
	config, err = rest.InClusterConfig()
	if err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			log.Fatal(err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	informerFactory := informers.NewSharedInformerFactory(clientSet, time.Hour*24)
	informer := informerFactory.Core().V1().Pods().Informer()
	if _, err = informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				enqueue(obj, queue)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				enqueue(newObj, queue)
			},
			DeleteFunc: func(obj interface{}) {
				enqueue(obj, queue)
			},
		},
	); err != nil {
		log.Fatal(err)
	}

	stop := make(chan struct{})
	defer close(stop)

	informerFactory.Start(stop)
	informerFactory.WaitForCacheSync(stop)

	controller := NewController(queue, informer.GetIndexer(), informer)
	go func() {
		for {
			if !controller.processNextItem() {
				return
			}
		}
	}()

	<-stop
}

func enqueue(obj any, queue workqueue.TypedRateLimitingInterface[string]) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Println(err)
		return
	}
	queue.Add(key)

}

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.TypedRateLimitingInterface[string]
	informer cache.Controller
}

func NewController(
	queue workqueue.TypedRateLimitingInterface[string],
	indexer cache.Indexer,
	informer cache.Controller,
) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncToStdout(key)
	c.handleErr(err, key)
	return true
}

// 输出日志
func (c *Controller) syncToStdout(key string) error {
	// 通过 key 从 indexer 中获取完整的对象
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Deployment %s does not exist anymore\n", key)
	} else {
		pod := obj.(*v1.Pod)
		fmt.Printf("Add/Delete for Pod %s, Namespace: %s\n", pod.Name, pod.Namespace)
	}
	return nil
}

func (c *Controller) handleErr(err error, key string) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 6 {
		fmt.Printf("Retry %d for key %s\n", c.queue.NumRequeues(key), key)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	fmt.Printf("Dropping deployment %q out of the queue: %v\n", key, err)
}

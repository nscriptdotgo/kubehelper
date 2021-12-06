package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/noahjd/kube-scaler/pkg/k8sapiconn"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

type ScaleRequest struct {
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
	Replicas   int32  `json:"replicas"`
}

func int32Ptr(i int32) *int32 { return &i }

func scaleDeployment(w http.ResponseWriter, r *http.Request) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	var deploymentsClient = clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	var patchUpdateObject ScaleRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Fprintf(w, "Please ensure your request is properly formed.")
	}
	json.Unmarshal(reqBody, &patchUpdateObject)

	fmt.Print(patchUpdateObject.Deployment)
	if err != nil {
		fmt.Fprintf(w, "Something happened during unmarshaling")
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(context.TODO(), patchUpdateObject.Deployment, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = int32Ptr(patchUpdateObject.Replicas)
		_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func handleRequests() {
	http.HandleFunc("/api/v1/deployment/scale", scaleDeployment)
	// http.HandleFunc("/api/v1/deployment/list", getDeployment).Methods("GET")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	handleRequests()
}

/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	logv1 "github.com/sawyer523/llm-rag-log-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// RagLogPilotReconciler reconciles a RagLogPilot object
type RagLogPilotReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	KubeClient *kubernetes.Clientset
}

// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RagLogPilot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *RagLogPilotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ragLogPilot logv1.RagLogPilot
	if err := r.Get(ctx, req.NamespacedName, &ragLogPilot); err != nil {
		logger.Error(err, "unable to fetch RagLogPilot")
		return ctrl.Result{}, err
	}

	if ragLogPilot.Status.ConversationId == "" {
		logger.Info("Creating a new conversation")
		conversationId, err := r.createNewConversation(ctx, ragLogPilot)
		if err != nil {
			logger.Error(err, "unable to create a new conversation")
			return ctrl.Result{}, err
		}

		ragLogPilot.Status.ConversationId = conversationId
		if err = r.Status().Update(ctx, &ragLogPilot); err != nil {
			logger.Error(err, "unable to update RagLogPilot status")
			return ctrl.Result{}, err
		}
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(ragLogPilot.Spec.WorkloadNameSpace)); err != nil {
		logger.Error(err, "unable to list pods")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		logString, err := r.getPodLogs(ctx, pod)
		if err != nil {
			logger.Error(err, "unable to get pod logs")
			continue
		}
		var errorLog []string
		logLines := strings.Split(logString, "\n")
		for _, line := range logLines {
			if strings.Contains(line, "ERROR") {
				errorLog = append(errorLog, line)
			}
		}

		if len(errorLog) > 0 {
			combinedErrLog := strings.Join(errorLog, "\n")
			fmt.Println(combinedErrLog)

			// 调用 Ragflow API
			answer, err := r.queryRagSystem(combinedErrLog, ragLogPilot)
			if err != nil {
				continue
			}

			if err = r.sendFeishuAlert(
				"https://open.feishu.cn/open-apis/bot/v2/hook/d5e267dc-a92f-43d3-bc45-106b5e718c49",
				answer,
			); err != nil {
				logger.Error(err, "unable to send Feishu alert")
			}
			logger.Info("RAG system response", "answer", answer)
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RagLogPilotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var (
		kubeConfig *string
		config     *rest.Config
	)

	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String(
			"kubeConfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeConfig file",
		)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig); err != nil {
			return err
		}
	}

	r.KubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&logv1.RagLogPilot{}).
		Complete(r)
}

func (r *RagLogPilotReconciler) queryRagSystem(podLog string, ragLogPilot logv1.RagLogPilot) (string, error) {
	payload := map[string]any{
		"conversation_id": ragLogPilot.Status.ConversationId,
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf(
					"以下是获取到的日志：%s，请基于运维知识库进行解答，如果你不知道，就说不知道",
					podLog,
				),
			},
		},
		"stream": false,
	}

	body, _ := json.Marshal(payload)
	u, err := getURL(ragLogPilot, "completion")
	if err != nil {
		return "", err
	}
	req, err := getRagReq(ragLogPilot, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var result map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["data"].(map[string]any)["answer"].(string), nil

}

func (r *RagLogPilotReconciler) getPodLogs(
	ctx context.Context,
	pod corev1.Pod,
) (string, error) {
	tailLines := int64(20)
	logOptions := &corev1.PodLogOptions{TailLines: &tailLines}
	req := r.KubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	logStream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}

	defer logStream.Close()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(logStream); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *RagLogPilotReconciler) createNewConversation(
	ctx context.Context,
	pilot logv1.RagLogPilot,
) (string, error) {
	u, err := getURL(pilot, "new_conversation")
	if err != nil {
		return "", err
	}
	query := &url.Values{}
	query.Add("user_id", uuid.New().String())
	u.RawQuery = query.Encode()
	req, err := getRagReq(pilot, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var result map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil
	}

	return result["data"].(map[string]any)["answer"].(string), nil
}

// sendFeishuAlert 发送飞书告警
func (r *RagLogPilotReconciler) sendFeishuAlert(webhook, analysis string) error {
	// 飞书消息内容
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": analysis,
		},
	}

	// 将消息内容序列化为 JSON
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(messageBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发出请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send Feishu alert, status code: %d", resp.StatusCode)
	}

	return nil
}

func getURL(pilot logv1.RagLogPilot, elem ...string) (*url.URL, error) {
	u, err := url.Parse(pilot.Spec.RagFlowEndpoint)
	if err != nil {
		return nil, err
	}

	return u.JoinPath(elem...), nil
}

func getRagReq(pilot logv1.RagLogPilot, method string, u *url.URL, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+pilot.Spec.RagFlowToken)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

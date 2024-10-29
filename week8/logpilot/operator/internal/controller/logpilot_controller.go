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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ollamaapi "github.com/ollama/ollama/api"
	openai "github.com/sashabaranov/go-openai"
	logv1 "github.com/sawyer523/llm-log-operator/api/v1"
)

// LogPilotReconciler reconciles a LogPilot object
type LogPilotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=log.aiops.com,resources=logpilots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=log.aiops.com,resources=logpilots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=log.aiops.com,resources=logpilots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LogPilot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *LogPilotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var logPilot logv1.LogPilot
	if err := r.Get(ctx, req.NamespacedName, &logPilot); err != nil {
		logger.Error(err, "unable to fetch LogPilot")
		return ctrl.Result{}, err
	}

	currentTime := time.Now().Unix()
	preTimeStamp := logPilot.Status.PreTimeStamp

	var preTime int64
	if preTimeStamp == "" {
		preTime = currentTime - 5
	} else {
		preTime, _ = strconv.ParseInt(preTimeStamp, 10, 64)
	}

	lokiQuery := logPilot.Spec.LokiPromQL
	endTime := currentTime * int64(time.Second)
	startTime := (preTime - 5) * int64(time.Second)

	if endTime <= startTime {
		logger.Info("endTime <= startTime")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	startTimeForUpdate := currentTime
	lokiURL := fmt.Sprintf(
		"%s/loki/api/v1/query_range?query=%s&start=%d&end=%d",
		logPilot.Spec.LokiURL,
		url.QueryEscape(lokiQuery),
		startTime,
		endTime,
	)

	lokiLogs, err := r.queryLoki(lokiURL)
	if err != nil {
		logger.Error(err, "query loki error")
		return ctrl.Result{}, err
	}

	if lokiLogs != "" {
		fmt.Println("send logs to llm")
		resp, err := r.analyzeLogsWithLLM(
			logPilot.Spec.LokiURL,
			logPilot.Spec.LLMToken,
			logPilot.Spec.LLMModel,
			logPilot.Spec.LLMType,
			lokiLogs,
		)

		if err != nil {
			logger.Error(err, "analyze logs with llm error")
			return ctrl.Result{}, err
		}

		if resp.HasErrors {
			err := r.sendFeishuAlert(logPilot.Spec.FeishuWeebhook, resp.Analysis)
			if err != nil {
				logger.Error(err, "send alert to feishu error")
				return ctrl.Result{}, err
			}
		}
	}

	logPilot.Status.PreTimeStamp = fmt.Sprintf("%d", startTimeForUpdate)
	if err := r.Status().Update(ctx, &logPilot); err != nil {
		logger.Error(err, "update status error")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPilotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logv1.LogPilot{}).
		Complete(r)
}

func (r *LogPilotReconciler) queryLoki(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var lokiResp map[string]any
	if err = json.Unmarshal(body, &lokiResp); err != nil {
		return "", err
	}

	data, ok := lokiResp["data"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("data not found")
	}

	result, ok := data["result"].([]any)
	if !ok || len(result) == 0 {
		return "", nil
	}

	return string(body), nil

}

type LLMAnalysisResult struct {
	HasErrors bool
	Analysis  string
}

func (r *LogPilotReconciler) analyzeLogsWithLLM(
	endpoint,
	token,
	model,
	tye,
	logs string,
) (*LLMAnalysisResult, error) {
	var result *LLMAnalysisResult
	content := fmt.Sprintf(
		"你现在是一名运维专家，以下日志是从日志系统里获取的日志，请分析日志的错误等级，如果遇到严重的问题，例如请求外部系统失败、外部系统故障、致命故障、数据库连接错误等严重问题时，请给出简短的建议，对于你认为严重需要通知运营人员的，请在返回内容里增加[feishu]标识:\n%s",
		logs,
	)
	switch tye {
	case "openapi":
		config := openai.DefaultConfig(token)
		config.BaseURL = endpoint
		client := openai.NewClientWithConfig(config)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: content,
					},
				},
			},
		)
		if err != nil {
			return nil, err
		}
		result = parseLLMResponse(resp.Choices[0].Message.Content)
	case "ollama":
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		client := ollamaapi.NewClient(u, http.DefaultClient)
		req := &ollamaapi.ChatRequest{
			Model: model,
			Messages: []ollamaapi.Message{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
			Stream: new(bool),
		}
		*req.Stream = false

		if err = client.Chat(
			context.Background(),
			req,
			func(response ollamaapi.ChatResponse) error {
				result = parseLLMResponse(response.Message.Content)
				return nil
			},
		); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *LogPilotReconciler) sendFeishuAlert(webhook, content string) error {
	// 飞书消息内容
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
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

// parseLLMResponse 解析 LLM API 的响应
func parseLLMResponse(content string) *LLMAnalysisResult {
	result := &LLMAnalysisResult{
		Analysis: content, // 从 LLM 返回的文本中获取分析结果
	}

	// 简单判断分析结果是否包含错误的标识符
	if strings.Contains(strings.ToLower(result.Analysis), "feishu") {
		result.HasErrors = true
	} else {
		result.HasErrors = false
	}

	return result
}

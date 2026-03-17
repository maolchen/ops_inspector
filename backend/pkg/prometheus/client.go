package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// QueryResult 查询结果
type QueryResult struct {
	Instance string  `json:"instance"`
	Value    float64 `json:"value"`
	Labels   string  `json:"labels"`
}

// Client Prometheus 客户端
type Client struct {
	httpClient *http.Client
}

// NewClient 创建客户端（支持HTTPS自签证书）
func NewClient() *Client {
	// 创建自定义Transport，跳过TLS证书验证
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略证书验证，支持自签证书
		},
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Query 即时查询
func (c *Client) Query(address, token, query string) ([]QueryResult, error) {
	// 打印查询日志
	log.Printf("[Prometheus Query] address=%s, query=%s", address, query)

	// 构建查询 URL
	u, err := url.Parse(address)
	if err != nil {
		log.Printf("[Prometheus Query Error] parse url failed: %v", err)
		return nil, err
	}

	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Printf("[Prometheus Query Error] create request failed: %v", err)
		return nil, err
	}

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Prometheus Query Error] send request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		log.Printf("[Prometheus Query Error] decode response failed: %v", err)
		return nil, err
	}

	if promResp.Status != "success" {
		log.Printf("[Prometheus Query Error] query failed: %s", promResp.Error)
		return nil, fmt.Errorf("prometheus query failed: %s", promResp.Error)
	}

	// 转换结果
	var results []QueryResult
	for _, r := range promResp.Data.Result {
		instance := r.Metric["instance"]
		if instance == "" {
			instance = r.Metric["node"]
		}
		if instance == "" {
			instance = r.Metric["pod"]
		}

		var value float64
		if len(r.Value) >= 2 {
			switch v := r.Value[1].(type) {
			case string:
				fmt.Sscanf(v, "%f", &value)
			case float64:
				value = v
			}
		}

		labelsJSON, _ := json.Marshal(r.Metric)

		results = append(results, QueryResult{
			Instance: instance,
			Value:    value,
			Labels:   string(labelsJSON),
		})
	}

	// 打印结果日志
	log.Printf("[Prometheus Query Result] query=%s, count=%d", query, len(results))
	for i, r := range results {
		log.Printf("[Prometheus Query Result] [%d] instance=%s, value=%.2f, labels=%s", i, r.Instance, r.Value, r.Labels)
	}

	return results, nil
}

// QueryRange 范围查询（趋势数据）
func (c *Client) QueryRange(address, token, query string) (string, error) {
	// 构建查询 URL
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}

	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", time.Now().Add(-7*24*time.Hour).Unix()))
	q.Set("end", fmt.Sprintf("%d", time.Now().Unix()))
	q.Set("step", "5m")
	u.RawQuery = q.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// 添加认证头
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(address, token string) error {
	u, err := url.Parse(address)
	if err != nil {
		return err
	}

	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", "up")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	return nil
}

// QueryRaw 执行原始查询（用于测试规则）
func (c *Client) QueryRaw(address, token, query string) (interface{}, error) {
	results, err := c.Query(address, token, query)
	if err != nil {
		return nil, err
	}
	return results, nil
}

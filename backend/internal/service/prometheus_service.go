package service

import (
	"ops-inspection/pkg/prometheus"
)

type PrometheusService struct {
	client *prometheus.Client
}

func NewPrometheusService() *PrometheusService {
	return &PrometheusService{
		client: prometheus.NewClient(),
	}
}

func (s *PrometheusService) Query(address, token, query string) ([]prometheus.QueryResult, error) {
	return s.client.Query(address, token, query)
}

func (s *PrometheusService) QueryRange(address, token, query string) (string, error) {
	return s.client.QueryRange(address, token, query)
}

func (s *PrometheusService) TestConnection(address, token string) error {
	return s.client.TestConnection(address, token)
}

func (s *PrometheusService) TestQuery(address, token, query string) (interface{}, error) {
	return s.client.QueryRaw(address, token, query)
}

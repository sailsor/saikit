package supermng

import (
	"fmt"
	"github.com/abrander/go-supervisord"
)

type SuperClient struct {
	client *supervisord.Client
}

func NewSuperClient(superConfig *SuperConfig) (*SuperClient, error) {
	url := fmt.Sprintf("http://%s:%s/RPC2", superConfig.Host, superConfig.Port)
	client, err := supervisord.NewClient(url,
		supervisord.WithAuthentication(superConfig.User, superConfig.Password))
	if err != nil {
		return nil, err
	}

	return &SuperClient{client: client}, nil
}

func (s *SuperClient) StartProcess(name string) error {
	return s.client.StartProcess(name, true)
}

func (s *SuperClient) StopProcess(name string) error {
	return s.client.StopProcess(name, true)
}

func (s *SuperClient) Restart(name string) error {
	_ = s.client.StopProcess(name, true)
	err := s.client.StartProcess(name, true)

	return err
}

func (s *SuperClient) StartAllProcesses() ([]supervisord.ProcessInfo, error) {
	return s.client.StartAllProcesses(true)
}

func (s *SuperClient) StopAllProcesses() error {
	_, err := s.client.StopAllProcesses(true)
	return err
}

func (s *SuperClient) RestartALL() error {
	_, _ = s.client.StopAllProcesses(true)
	_, err := s.client.StartAllProcesses(true)

	return err
}

func (s *SuperClient) GetProcessInfo(name string) (*supervisord.ProcessInfo, error) {
	return s.client.GetProcessInfo(name)
}

func (s *SuperClient) GetAllProcessInfo() ([]supervisord.ProcessInfo, error) {
	return s.client.GetAllProcessInfo()
}

/*
ReloadConfig
supervisorctl reread
*/
func (s *SuperClient) ReloadConfig() (added, changed, removed []string, err error) {
	added, changed, removed, err = s.client.ReloadConfig()
	return
}

/*
Update
supervisorctl update
*/
func (s *SuperClient) Update() error {
	return s.client.Update()
}

func (s *SuperClient) Close() error {
	return s.client.Close()
}

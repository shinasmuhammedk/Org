package executor

import (
	"encoding/json"
	"time"
)

func (e *Executor) executeDelay(
	config []byte,
) ([]byte, error) {

	var cfg struct {
		Duration int `json:"duration"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}

	time.Sleep(time.Duration(cfg.Duration) * time.Second)

	return []byte(`{"status":"completed"}`), nil
}
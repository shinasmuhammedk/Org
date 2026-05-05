package executor

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (e *Executor) executeHTTPRequest(config []byte) error {

	var cfg struct {
		Method string                 `json:"method"`
		URL    string                 `json:"url"`
		Body   map[string]interface{} `json:"body"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil{
        return err
    }
    
    bodyBytes, _ := json.Marshal(cfg.Body)
    
    req,err := http.NewRequest(cfg.Method, cfg.URL, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    
    defer resp.Body.Close()
    
    return nil
}

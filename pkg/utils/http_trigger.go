package utils

import (
	"fmt"
	"github.com/go-logr/logr"
	"os"
	"strings"
)

func statusChangedTrigger(target string, instanceName string, namespace string) error {
	params := make(map[string]string, 1)
	params["instanceName"] = instanceName
	params["namespace"] = namespace
	resp, err := HTTPPut(target, nil, params, nil)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("invalid trigger request to target %s, code is %d", target, resp.StatusCode)
	}
	return nil
}

func StatusChangedTrigger(instanceName string, namespace string, log logr.Logger) {
	if targets := os.Getenv("STATUS_CHANGED_TRIGGER"); len(targets) != 0 {
		targetList := strings.Split(targets, ",")
		for _, target := range targetList {
			go func(target string) {
				err := statusChangedTrigger(target, instanceName, namespace)
				if err != nil {
					log.Error(err, "Status changed trigger start failed", "trigger.target", target, "instance.Name", instanceName)
				} else {
					log.Info("Triggered by status changed", "trigger.target", target, "instance.Name", instanceName)
				}
			}(strings.TrimSpace(target))
		}
	}
}

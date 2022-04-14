package utils

import "fmt"

func StatusChangedTrigger(target string, instanceName string) error {
	params := make(map[string]string, 1)
	params["instanceName"] = instanceName
	resp, err := HTTPPut(target, nil, params, nil)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("invalid trigger request, code is %d", resp.StatusCode)
	}
	return nil
}

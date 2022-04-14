package utils

func StatusChangedTrigger(target string, instanceName string) error {
	params := make(map[string]string, 1)
	params["instanceName"] = instanceName
	_, err := HTTPGet(target, params, nil)
	return err
}

package mcapi

func (c *Client) CreateProcess(projectID, experimentID, name, processType string, setups []Setup) (*Process, error) {
	var result struct {
		Data Process `json:"data"`
	}

	if setups == nil {
		setups = make([]Setup, 0)
	}

	body := struct {
		ProjectID    string  `json:"project_id"`
		ExperimentID string  `json:"experiment_id"`
		Name         string  `json:"name"`
		ProcessType  string  `json:"process_type"`
		Attributes   []Setup `json:"attributes"`
	}{
		ProjectID:    projectID,
		ExperimentID: experimentID,
		Name:         name,
		Attributes:   setups,
		ProcessType:  processType,
	}

	if err := c.post(&result, body, "createProcess"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

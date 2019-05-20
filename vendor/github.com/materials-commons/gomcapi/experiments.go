package mcapi

func (c *Client) CreateExperiment(projectID, name, description string, inProgress bool) (*Experiment, error) {
	var result struct {
		Data Experiment `json:"data"`
	}

	body := map[string]interface{}{
		"project_id":  projectID,
		"name":        name,
		"description": description,
		"in_progress": inProgress,
	}

	if err := c.post(&result, body, "createExperimentInProject"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateExperimentProgressStatus(projectID, experimentID string, inProgress bool) error {
	var result struct {
		Data struct {
			Success bool `json:"success"`
		} `json:"data"`
	}

	body := map[string]interface{}{
		"project_id":    projectID,
		"experiment_id": experimentID,
		"in_progress":   inProgress,
	}

	return c.post(&result, body, "updateExperimentProgressStatus")
}

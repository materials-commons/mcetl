package mcapi

func (c *Client) CreateProject(name, description string) (*Project, error) {
	body := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		Name:        name,
		Description: description,
	}

	var result struct {
		Data Project `json:"data"`
	}

	if err := c.post(&result, body, "createProject"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetProjectOverviewByName(name string) (*Project, error) {
	body := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	var result struct {
		Data Project `json:"data"`
	}

	if err := c.post(&result, body, "getProjectOverviewByName"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteProject(projectID string) error {
	body := map[string]interface{}{"project_id": projectID}

	var result struct {
		Data struct {
			Success string `json:"success"`
		} `json:"data"`
	}

	return c.post(&result, body, "deleteProject")
}

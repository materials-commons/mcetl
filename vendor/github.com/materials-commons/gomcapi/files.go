package mcapi

func (c *Client) GetFileByPathInProject(filePath, projectID string) (*File, error) {
	var result struct {
		Data File `json:"data"`
	}

	body := struct {
		ProjectID string `json:"project_id"`
		Path      string `json:"path"`
	}{
		ProjectID: projectID,
		Path:      filePath,
	}

	if err := c.post(&result, body, "getFileByPathInProject"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

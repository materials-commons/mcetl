package mcapi

func (c *Client) CreateSample(projectID, experimentID, name string, attributes []Property) (*Sample, error) {
	var result struct {
		Data Sample `json:"data"`
	}

	if attributes == nil {
		attributes = make([]Property, 0)
	}

	body := struct {
		ProjectID    string     `json:"project_id"`
		ExperimentID string     `json:"experiment_id"`
		Name         string     `json:"name"`
		Attributes   []Property `json:"attributes"`
	}{
		ProjectID:    projectID,
		ExperimentID: experimentID,
		Name:         name,
		Attributes:   attributes,
	}

	if err := c.post(&result, body, "createSample"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type ConnectSampleToProcess struct {
	ProcessID     string
	SampleID      string
	PropertySetID string
	Transform     bool
}

func (c *Client) AddSampleToProcess(projectID, experimentID string, simple bool, connect ConnectSampleToProcess) (*Sample, error) {
	var result struct {
		Data Sample `json:"data"`
	}

	body := struct {
		ProjectID        string `json:"project_id"`
		ExperimentID     string `json:"experiment_id"`
		ProcessID        string `json:"process_id"`
		SampleID         string `json:"sample_id"`
		PropertySetID    string `json:"property_set_id"`
		Transform        bool   `json:"transform"`
		ReturnFullSample bool   `json:"return_full_sample"`
	}{
		ProjectID:        projectID,
		ExperimentID:     experimentID,
		ProcessID:        connect.ProcessID,
		SampleID:         connect.SampleID,
		PropertySetID:    connect.PropertySetID,
		Transform:        connect.Transform,
		ReturnFullSample: simple,
	}

	if err := c.post(&result, body, "addSampleToProcess"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type SampleToConnect struct {
	SampleID      string `json:"sample_id"`
	PropertySetID string `json:"property_set_id"`
	Name          string `json:"name"`
}

type ConnectSamplesToProcess struct {
	ProcessID string
	Transform bool
	Samples   []SampleToConnect
}

func (c *Client) AddSamplesToProcess(projectID, experimentID string, connect ConnectSamplesToProcess) ([]Sample, error) {
	var result struct {
		Data []Sample `json:"data"`
	}

	body := struct {
		ProjectID    string            `json:"project_id"`
		ExperimentID string            `json:"experiment_id"`
		ProcessID    string            `json:"process_id"`
		Transform    bool              `json:"transform"`
		Samples      []SampleToConnect `json:"samples"`
	}{
		ProjectID:    projectID,
		ExperimentID: experimentID,
		ProcessID:    connect.ProcessID,
		Transform:    connect.Transform,
		Samples:      connect.Samples,
	}

	if err := c.post(&result, body, "addSamplesToProcess"); err != nil {
		return nil, err
	}

	return result.Data, nil
}

type ConnectSampleAndFilesToProcess struct {
	ProcessID     string
	SampleID      string
	PropertySetID string
	Transform     bool
	FilesByName   []FileAndDirection
	FilesByID     []FileAndDirection
}

type FileAndDirection struct {
	FileID    string `json:"file_id,omitempty"`
	Path      string `json:"path,omitempty"`
	Direction string `json:"direction"`
}

func (c *Client) AddSampleAndFilesToProcess(projectID, experimentID string, simple bool, connect ConnectSampleAndFilesToProcess) (*Sample, error) {
	var result struct {
		Data Sample `json:"data"`
	}

	body := struct {
		ProjectID        string             `json:"project_id"`
		ExperimentID     string             `json:"experiment_id"`
		ProcessID        string             `json:"process_id"`
		SampleID         string             `json:"sample_id"`
		PropertySetID    string             `json:"property_set_id"`
		Transform        bool               `json:"transform"`
		FilesByName      []FileAndDirection `json:"files_by_name,omitempty"`
		FilesByID        []FileAndDirection `json:"files_by_id,omitempty"`
		ReturnFullSample bool               `json:"return_full_sample"`
	}{
		ProjectID:        projectID,
		ExperimentID:     experimentID,
		ProcessID:        connect.ProcessID,
		SampleID:         connect.SampleID,
		PropertySetID:    connect.PropertySetID,
		Transform:        connect.Transform,
		ReturnFullSample: simple,
	}

	if len(connect.FilesByName) != 0 {
		body.FilesByName = connect.FilesByName
	}

	if len(connect.FilesByID) != 0 {
		body.FilesByID = connect.FilesByID
	}

	if err := c.post(&result, body, "addSampleAndFilesToProcess"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type SampleProperty struct {
	Name         string                 `json:"name"`
	ID           string                 `json:"id,omitempty"`
	Measurements []Measurement          `json:"measurements"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type SampleMeasurements struct {
	SampleID      string
	PropertySetID string
	Attributes    []SampleProperty
}

func (c *Client) AddMeasurementsToSampleInProcess(projectID, experimentID, processID string, simple bool, sm SampleMeasurements) (*Sample, error) {
	var result struct {
		Data Sample `json:"data"`
	}

	body := struct {
		ProjectID        string           `json:"project_id"`
		ExperimentID     string           `json:"experiment_id"`
		ProcessID        string           `json:"process_id"`
		SampleID         string           `json:"sample_id"`
		PropertySetID    string           `json:"property_set_id"`
		Attributes       []SampleProperty `json:"attributes"`
		ReturnFullSample bool             `json:"return_full_sample"`
	}{
		ProjectID:        projectID,
		ExperimentID:     experimentID,
		ProcessID:        processID,
		SampleID:         sm.SampleID,
		PropertySetID:    sm.PropertySetID,
		Attributes:       sm.Attributes,
		ReturnFullSample: simple,
	}

	if body.Attributes == nil {
		body.Attributes = make([]SampleProperty, 0)
	}

	for _, attr := range body.Attributes {
		if attr.Measurements == nil {
			attr.Measurements = make([]Measurement, 0)
		}

		if attr.Metadata == nil {
			attr.Metadata = make(map[string]interface{}, 0)
		}
	}

	if err := c.post(&result, body, "addMeasurementsToSampleInProcess"); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

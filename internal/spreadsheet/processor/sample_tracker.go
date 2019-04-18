package processor

import mcapi "github.com/materials-commons/gomcapi"

type sampleTracker struct {
	// Track all samples that have been created, this is the samples
	// initial starting state (came from a Create Samples process).
	createdSamples map[string]*mcapi.Sample

	// Track all samples associated with a particular process id,
	// if the sample hasn't been associated with a process then
	// we can check if it has been created, and if not create it,
	// then associate the sample with the process.
	samplesInProcess map[string][]*mcapi.Sample
}

func newSampleTracker() *sampleTracker {
	return &sampleTracker{
		createdSamples:   make(map[string]*mcapi.Sample),
		samplesInProcess: make(map[string][]*mcapi.Sample),
	}
}

func (t *sampleTracker) findCreatedSample(sampleName string) *mcapi.Sample {
	return t.createdSamples[sampleName]
}

func (t *sampleTracker) addCreatedSample(sample *mcapi.Sample) {
	t.createdSamples[sample.Name] = sample
}

func (t *sampleTracker) findSampleByProcessID(sampleName, processID string) *mcapi.Sample {
	samples := t.samplesInProcess[processID]

	for _, sample := range samples {
		if sample.Name == sampleName {
			return sample
		}
	}

	return nil
}

func (t *sampleTracker) addSampleByProcessID(sample *mcapi.Sample, processID string) {
	t.samplesInProcess[processID] = append(t.samplesInProcess[processID], sample)
}

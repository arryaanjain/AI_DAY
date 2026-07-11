package worker
// Job payloads always carry jobId, attempt and traceId. Concrete queue consumers
// are added after Redis/S3 provider credentials are configured.
type Payload struct { JobID string `json:"jobId"`; Attempt int `json:"attempt"`; TraceID string `json:"traceId"` }

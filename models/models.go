package models

type Payload struct {
	Caller string `json:"caller"`
	Done   bool   `json:"done"`
	Passed bool   `json:"passed"`
}

type Source struct {
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Path            string `json:"path"`
	Endpoint        string `json:"endpoint"`
	TLSSkipVerify   bool   `json:"tls_skip_verify"`
}

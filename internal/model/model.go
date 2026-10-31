package model

import "time"

// Model is a registered inference model and its desired state.
type Model struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Image           string    `json:"image"`
	HFRepo          string    `json:"hf_repo"`
	DesiredReplicas int       `json:"desired_replicas"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateModelInput struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	HFRepo          string `json:"hf_repo"`
	DesiredReplicas int    `json:"desired_replicas"`
}

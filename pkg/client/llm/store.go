package llm

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Store struct {
	endpoints []LLMEndpoint
}

func seedEndpoints() []LLMEndpoint {
	const seedTimestamp = "2024-01-01T00:00:00Z"
	return []LLMEndpoint{
		{
			EndpointID:       "llm-ep-001",
			EndpointName:     "chat-prod",
			Model:            "GLM-4.7",
			Project:          "project-alpha",
			Status:           StatusRunning,
			CreatedTimestamp: seedTimestamp,
			EndpointAddress:  "https://llm-ep-001.example.internal",
		},
		{
			EndpointID:       "llm-ep-002",
			EndpointName:     "code-assist",
			Model:            "GPT OSS 120b",
			Project:          "project-beta",
			Status:           StatusRunning,
			CreatedTimestamp: seedTimestamp,
			EndpointAddress:  "https://llm-ep-002.example.internal",
		},
		{
			EndpointID:       "llm-ep-003",
			EndpointName:     "summarizer",
			Model:            "GLM-4.6",
			Project:          "project-alpha",
			Status:           StatusRunning,
			CreatedTimestamp: seedTimestamp,
			EndpointAddress:  "https://llm-ep-003.example.internal",
		},
	}
}

func NewStore() *Store {
	return &Store{endpoints: seedEndpoints()}
}

func (s *Store) List() ([]LLMEndpoint, error) {
	out := make([]LLMEndpoint, len(s.endpoints))
	copy(out, s.endpoints)
	return out, nil
}

func (s *Store) Create(name, model, project string) (*LLMEndpoint, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateModel(model); err != nil {
		return nil, err
	}

	endpoint := LLMEndpoint{
		EndpointID:       s.newEndpointID(),
		EndpointName:     name,
		Model:            model,
		Project:          project,
		Status:           StatusPending,
		CreatedTimestamp: time.Now().UTC().Format(time.RFC3339),
		EndpointAddress:  "",
	}

	s.endpoints = append(s.endpoints, endpoint)

	created := endpoint
	return &created, nil
}

func (s *Store) Delete(endpointID, project string) error {
	for i, e := range s.endpoints {
		if e.EndpointID == endpointID && e.Project == project {
			s.endpoints = append(s.endpoints[:i], s.endpoints[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("endpoint %q not found in project %q", endpointID, project)
}

func (s *Store) newEndpointID() string {
	for {
		id := "llm-ep-" + randomHex(6)
		if !s.hasEndpointID(id) {
			return id
		}
	}
}

func (s *Store) hasEndpointID(id string) bool {
	for _, e := range s.endpoints {
		if e.EndpointID == id {
			return true
		}
	}
	return false
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

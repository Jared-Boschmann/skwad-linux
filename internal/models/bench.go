package models

import "github.com/google/uuid"

// BenchAgent is a saved agent configuration template.
type BenchAgent struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Avatar       string     `json:"avatar"`
	Folder       string     `json:"folder"`
	AgentType    AgentType  `json:"agentType"`
	ShellCommand string     `json:"shellCommand,omitempty"`
	PersonaID    *uuid.UUID `json:"personaId,omitempty"`
}

// ToAgent creates a new Agent from this bench template.
func (b *BenchAgent) ToAgent() *Agent {
	return &Agent{
		ID:           uuid.New(),
		Name:         b.Name,
		Avatar:       b.Avatar,
		Folder:       b.Folder,
		AgentType:    b.AgentType,
		ShellCommand: b.ShellCommand,
		PersonaID:    b.PersonaID,
	}
}

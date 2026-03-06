// Package notifications sends desktop notifications via libnotify (notify-send).
package notifications

import (
	"os/exec"
)

// Service sends desktop notifications via libnotify (notify-send).
type Service struct {
	enabled bool
	appName string
}

// NewService creates a notification service.
func NewService(appName string, enabled bool) *Service {
	return &Service{appName: appName, enabled: enabled}
}

// SetEnabled toggles notifications.
func (s *Service) SetEnabled(v bool) { s.enabled = v }

// Notify sends a desktop notification with title and body.
func (s *Service) Notify(title, body string) {
	if !s.enabled {
		return
	}
	// Use notify-send (libnotify CLI) — available on most Linux desktops.
	// TODO: consider using a CGo binding to libnotify for richer control.
	go exec.Command("notify-send", "--app-name="+s.appName, title, body).Run()
}

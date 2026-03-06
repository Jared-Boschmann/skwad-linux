// Package voice provides push-to-talk voice input via OS speech recognition.
// The implementation is a stub pending selection of a speech-to-text backend.
package voice

// Service provides push-to-talk voice input via OS speech recognition.
//
// On Linux, speech recognition requires an external tool or library.
// Options:
//   - vosk (offline, via CGo binding or subprocess)
//   - whisper.cpp (offline, via subprocess)
//   - Google Speech-to-Text (online, via REST API)
//
// The service is a stub; the backend is selected at build time or runtime.
type Service struct {
	enabled       bool
	pushToTalkKey string
	autoInsert    bool
	recording     bool

	OnTranscription func(text string)
}

// NewService creates a voice input service.
func NewService(enabled bool, pushToTalkKey string, autoInsert bool) *Service {
	return &Service{
		enabled:       enabled,
		pushToTalkKey: pushToTalkKey,
		autoInsert:    autoInsert,
	}
}

// SetEnabled toggles voice input.
func (s *Service) SetEnabled(v bool) { s.enabled = v }

// StartRecording begins capturing audio.
func (s *Service) StartRecording() {
	if !s.enabled || s.recording {
		return
	}
	s.recording = true
	// TODO: start OS audio capture
}

// StopRecording stops capture and triggers transcription.
func (s *Service) StopRecording() {
	if !s.recording {
		return
	}
	s.recording = false
	// TODO: send audio to recogniser, call s.OnTranscription with result
}

// IsRecording returns true while audio is being captured.
func (s *Service) IsRecording() bool { return s.recording }

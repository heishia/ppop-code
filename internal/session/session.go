package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Model     string    `json:"model,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Manager struct {
	historyDir string
	maxHistory int
	current    *Session
}

func NewManager(historyDir string, maxHistory int) *Manager {
	return &Manager{
		historyDir: historyDir,
		maxHistory: maxHistory,
	}
}

func (m *Manager) NewSession(name string) *Session {
	session := &Session{
		ID:        fmt.Sprintf("session-%d", time.Now().UnixNano()),
		Name:      name,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.current = session
	return session
}

func (m *Manager) Current() *Session {
	if m.current == nil {
		m.current = m.NewSession("default")
	}
	return m.current
}

func (m *Manager) AddMessage(role, content, model string) {
	session := m.Current()
	session.Messages = append(session.Messages, Message{
		Role:      role,
		Content:   content,
		Model:     model,
		Timestamp: time.Now(),
	})
	session.UpdatedAt = time.Now()
}

func (m *Manager) Save() error {
	if m.current == nil {
		return nil
	}

	if err := os.MkdirAll(m.historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history dir: %w", err)
	}

	filename := filepath.Join(m.historyDir, m.current.ID+".json")
	data, err := json.MarshalIndent(m.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func (m *Manager) Load(sessionID string) (*Session, error) {
	filename := filepath.Join(m.historyDir, sessionID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	m.current = &session
	return &session, nil
}

func (m *Manager) List() ([]Session, error) {
	var sessions []Session

	entries, err := os.ReadDir(m.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return sessions, nil
		}
		return nil, fmt.Errorf("failed to read history dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filename := filepath.Join(m.historyDir, entry.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (m *Manager) Delete(sessionID string) error {
	filename := filepath.Join(m.historyDir, sessionID+".json")
	return os.Remove(filename)
}

func (m *Manager) Clear() {
	m.current = nil
}

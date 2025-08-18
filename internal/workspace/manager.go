package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/debug"
)

// Manager handles isolated workspace creation and management for teams
type Manager struct {
	baseDir    string
	workspaces map[string]*Workspace
	mutex      sync.RWMutex
}

// Workspace represents an isolated work environment for a team
type Workspace struct {
	ID          string
	Path        string
	TeamID      string
	CreatedAt   time.Time
	Description string
	mutex       sync.RWMutex
}

// NewManager creates a new workspace manager
func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "agentry_workspaces")
	}

	// Ensure base directory exists
	os.MkdirAll(baseDir, 0755)

	return &Manager{
		baseDir:    baseDir,
		workspaces: make(map[string]*Workspace),
	}
}

// CreateWorkspace creates a new isolated workspace for a team
func (m *Manager) CreateWorkspace(ctx context.Context, teamID, description string) (*Workspace, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := uuid.New().String()
	workspacePath := filepath.Join(m.baseDir, id)

	// Create workspace directory structure
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create common subdirectories
	subdirs := []string{"src", "docs", "tests", "output", "temp"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(workspacePath, subdir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create subdir %s: %w", subdir, err)
		}
	}

	// Create workspace info file
	infoPath := filepath.Join(workspacePath, ".agentry_workspace")
	infoContent := fmt.Sprintf("workspace_id: %s\nteam_id: %s\ndescription: %s\ncreated_at: %s\n",
		id, teamID, description, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create workspace info: %w", err)
	}

	workspace := &Workspace{
		ID:          id,
		Path:        workspacePath,
		TeamID:      teamID,
		CreatedAt:   time.Now(),
		Description: description,
	}

	m.workspaces[id] = workspace
	debug.Printf("Created workspace %s at %s for team %s", id, workspacePath, teamID)

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (m *Manager) GetWorkspace(id string) (*Workspace, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	workspace, exists := m.workspaces[id]
	return workspace, exists
}

// ListWorkspaces returns all active workspaces
func (m *Manager) ListWorkspaces() []*Workspace {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	workspaces := make([]*Workspace, 0, len(m.workspaces))
	for _, workspace := range m.workspaces {
		workspaces = append(workspaces, workspace)
	}
	return workspaces
}

// CleanupWorkspace removes a workspace and its contents
func (m *Manager) CleanupWorkspace(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	workspace, exists := m.workspaces[id]
	if !exists {
		return fmt.Errorf("workspace %s not found", id)
	}

	// Remove the workspace directory
	if err := os.RemoveAll(workspace.Path); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	delete(m.workspaces, id)
	debug.Printf("Cleaned up workspace %s", id)

	return nil
}

// GetWorkingDirectory returns the working directory path for the workspace
func (w *Workspace) GetWorkingDirectory() string {
	return w.Path
}

// CreateFile creates a file in the workspace
func (w *Workspace) CreateFile(relativePath, content string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	fullPath := filepath.Join(w.Path, relativePath)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// ReadFile reads a file from the workspace
func (w *Workspace) ReadFile(relativePath string) (string, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	fullPath := filepath.Join(w.Path, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ListFiles lists all files in the workspace
func (w *Workspace) ListFiles() ([]string, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var files []string
	err := filepath.Walk(w.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(w.Path, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})

	return files, err
}

// GetStats returns workspace statistics
func (w *Workspace) GetStats() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	stats := map[string]interface{}{
		"id":          w.ID,
		"team_id":     w.TeamID,
		"path":        w.Path,
		"created_at":  w.CreatedAt,
		"description": w.Description,
	}

	// Count files
	files, _ := w.ListFiles()
	stats["file_count"] = len(files)

	// Calculate total size
	var totalSize int64
	filepath.Walk(w.Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	stats["total_size"] = totalSize

	return stats
}

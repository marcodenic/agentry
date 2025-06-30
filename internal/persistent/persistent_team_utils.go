package persistent

import (
	"fmt"
	"net"
)

// findAvailablePort finds an available port in the configured range
func (pt *PersistentTeam) findAvailablePort() (int, error) {
	for port := pt.portRange.Start; port <= pt.portRange.End; port++ {
		if pt.isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", pt.portRange.Start, pt.portRange.End)
}

// isPortAvailable checks if a port is available for use
func (pt *PersistentTeam) isPortAvailable(port int) bool {
	// Check if any existing agent is using this port
	for _, agent := range pt.agents {
		if agent.Port == port {
			return false
		}
	}

	// Try to bind to the port to verify it's available
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	
	// Close the listener immediately
	listener.Close()
	return true
}



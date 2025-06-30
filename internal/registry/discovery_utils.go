package registry

// isExcluded checks if an agent ID is in the exclusion list
func (d *DiscoveryService) isExcluded(agentID string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if excluded == agentID {
			return true
		}
	}
	return false
}

// hasRequiredStatus checks if an agent has one of the required statuses
func (d *DiscoveryService) hasRequiredStatus(agentStatus AgentStatus, requiredStatuses []AgentStatus) bool {
	for _, required := range requiredStatuses {
		if agentStatus == required {
			return true
		}
	}
	return false
}

// hasAllCapabilities checks if an agent has all required capabilities
func (d *DiscoveryService) hasAllCapabilities(agentCaps, requiredCaps []string) bool {
	if len(requiredCaps) == 0 {
		return true
	}
	
	capMap := make(map[string]bool)
	for _, cap := range agentCaps {
		capMap[cap] = true
	}
	
	for _, required := range requiredCaps {
		if !capMap[required] {
			return false
		}
	}
	
	return true
}

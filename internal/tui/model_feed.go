package tui

func (m Model) applyFeedFilter() Model {
	if m.view.ActivityFeed == nil {
		return m
	}
	if !m.layout.feedFocused {
		m.view.ActivityFeed.ClearFilter()
		return m
	}
	if info, ok := m.infos[m.active]; ok && info != nil {
		name := info.Name
		if name == "" && info.Agent != nil {
			name = info.Agent.Role
			if name == "" {
				name = info.Agent.ModelName
			}
		}
		m.view.ActivityFeed.SetFilter(m.active, name)
	} else {
		m.view.ActivityFeed.ClearFilter()
	}
	return m
}

func (m Model) shouldShowFeed() bool {
	if m.view.ActivityFeed == nil {
		return false
	}
	if m.layout.feedFocused {
		return true
	}
	return m.view.ActivityFeed.HasEvents()
}

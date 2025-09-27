package team

import "github.com/marcodenic/agentry/internal/contracts"

// Ensure Team implements the public TeamService contract.
var _ contracts.TeamService = (*Team)(nil)

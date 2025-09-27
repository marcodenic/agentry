package team

import "context"

// Call delegates work to the named agent with enhanced communication logging.
func (t *Team) Call(ctx context.Context, agentID, input string) (string, error) {
	return newDelegationSession(t, agentID, input).Run(ctx)
}

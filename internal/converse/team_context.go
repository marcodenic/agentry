package converse

import (
	"context"

	"github.com/marcodenic/agentry/internal/team"
)

func contextWithTeam(ctx context.Context, t *Team) context.Context {
	return team.WithContext(ctx, t)
}

// TeamFromContext extracts a Team pointer if present.
func TeamFromContext(ctx context.Context) (*Team, bool) {
	caller, ok := team.FromContext(ctx)
	if !ok {
		return nil, false
	}
	t, ok := caller.(*Team)
	return t, ok
}

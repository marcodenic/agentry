package router

import (
	"strings"

	"github.com/yourname/agentry/internal/model"
)

type Selector interface {
	Select(prompt string) model.Client
}

type Rule struct {
	IfContains []string
	Client     model.Client
}

type Rules []Rule

func (r Rules) Select(p string) model.Client {
	for _, rule := range r {
		for _, sub := range rule.IfContains {
			if strings.Contains(strings.ToLower(p), strings.ToLower(sub)) {
				return rule.Client
			}
		}
	}
	return r[0].Client
}

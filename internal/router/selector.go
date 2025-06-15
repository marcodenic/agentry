package router

import (
	"strings"

	"github.com/marcodenic/agentry/internal/model"
)

type Selector interface {
	Select(prompt string) (model.Client, string)
}

type Rule struct {
	IfContains []string
	Name       string
	Client     model.Client
}

type Rules []Rule

func (r Rules) Select(p string) (model.Client, string) {
	for _, rule := range r {
		for _, sub := range rule.IfContains {
			if strings.Contains(strings.ToLower(p), strings.ToLower(sub)) {
				return rule.Client, rule.Name
			}
		}
	}
	return r[0].Client, r[0].Name
}

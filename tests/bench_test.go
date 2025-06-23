package tests

import (
    "testing"

    "github.com/marcodenic/agentry/internal/router"
    "github.com/marcodenic/agentry/internal/model"
)

func BenchmarkRulesSelect(b *testing.B) {
    rules := router.Rules{
        {IfContains: []string{"hello", "world"}, Client: model.NewMock()},
        {IfContains: []string{"foo"}, Client: model.NewMock()},
        {IfContains: []string{"bar"}, Client: model.NewMock()},
    }
    prompt := "say hello world"
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        rules.Select(prompt)
    }
}

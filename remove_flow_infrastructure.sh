#!/bin/bash

echo "Removing flow infrastructure from Agentry..."

# Remove flow command implementation
echo "Removing flow command implementation..."
rm -f cmd/agentry/flow_cmd.go

# Remove flow package
echo "Removing pkg/flow package..."
rm -rf pkg/flow/

# Remove flow-related test files
echo "Removing flow test files..."
rm -f tests/flow_parser_test.go
rm -f tests/flow_engine_test.go
rm -f tests/flow_tool_test.go

# Remove flow documentation
echo "Removing flow documentation..."
rm -f design/flow_dsl.md

# Remove flow examples
echo "Removing flow examples..."
rm -rf examples/flows/

# Remove flow configurations in teams
echo "Removing flow configurations..."
rm -f examples/teams/website/.agentry.flow.yaml
rm -f examples/teams/docs/.agentry.flow.yaml

echo "Flow infrastructure removal complete."
echo ""
echo "Files removed:"
echo "- cmd/agentry/flow_cmd.go"
echo "- pkg/flow/ (entire directory)"
echo "- tests/flow_parser_test.go"
echo "- tests/flow_engine_test.go"
echo "- tests/flow_tool_test.go"
echo "- design/flow_dsl.md"
echo "- examples/flows/ (entire directory)"
echo "- examples/teams/website/.agentry.flow.yaml"
echo "- examples/teams/docs/.agentry.flow.yaml"

#!/usr/bin/env python3
"""
Test script to debug the Responses API function calling issue
"""
import os
import json
from openai import OpenAI

client = OpenAI(api_key=os.getenv('OPENAI_API_KEY'))

print("Testing direct Responses API call...")

response = client.responses.create(
    model="gpt-4o-mini",
    instructions="""You are an agent coordinator. When asked to spawn a coder, call the agent tool with:
{"agent": "coder", "input": "help with coding task"}""",
    input="spawn a coder",
    tools=[{
        "type": "function",
        "name": "agent",
        "description": "Delegate work to another agent",
        "parameters": {
            "type": "object",
            "properties": {
                "agent": {"type": "string", "description": "Name of the agent to delegate to"},
                "input": {"type": "string", "description": "Task description or input for the agent"}
            },
            "required": []
        }
    }]
)

print(f"Response ID: {response.id}")
print(f"Output text: {response.output_text}")

# Print the full response to understand the structure
print("\n=== FULL RESPONSE ===")
print(json.dumps(response.model_dump(), indent=2))

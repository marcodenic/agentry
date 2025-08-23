#!/usr/bin/env python3
import json
import requests
import os

# Test both endpoints to see which one actually works

api_key = os.getenv('OPENAI_API_KEY')
if not api_key:
    print("OPENAI_API_KEY not set")
    exit(1)

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {api_key}'
}

# Test 1: /v1/responses (what agentry is using)
print("=== Testing /v1/responses endpoint ===")
responses_payload = {
    "model": "gpt-3.5-turbo",
    "input": [
        {
            "role": "user",
            "content": [{"type": "input_text", "text": "Hello"}]
        }
    ]
}

try:
    resp = requests.post('https://api.openai.com/v1/responses', 
                        headers=headers, 
                        json=responses_payload)
    print(f"Status: {resp.status_code}")
    print(f"Response: {resp.text[:500]}...")
except Exception as e:
    print(f"Error: {e}")

print("\n=== Testing /v1/chat/completions endpoint ===")
# Test 2: /v1/chat/completions (standard endpoint)
chat_payload = {
    "model": "gpt-3.5-turbo",
    "messages": [
        {"role": "user", "content": "Hello"}
    ]
}

try:
    resp = requests.post('https://api.openai.com/v1/chat/completions',
                        headers=headers,
                        json=chat_payload)
    print(f"Status: {resp.status_code}")
    print(f"Response: {resp.text[:500]}...")
except Exception as e:
    print(f"Error: {e}")

print("\n=== Testing /v1/responses with tools (CORRECTED FORMAT) ===")
# Test 3: /v1/responses with function calling - trying the correct format based on error
responses_tools_payload = {
    "model": "gpt-3.5-turbo",
    "input": [
        {
            "role": "user",
            "content": [{"type": "input_text", "text": "What's the weather like?"}]
        }
    ],
    "tools": [
        {
            "type": "function",
            "name": "get_weather",
            "description": "Get weather information",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"}
                }
            }
        }
    ],
    "tool_choice": "auto"
}

try:
    resp = requests.post('https://api.openai.com/v1/responses',
                        headers=headers,
                        json=responses_tools_payload)
    print(f"Status: {resp.status_code}")
    print(f"Full Response: {resp.text}")
except Exception as e:
    print(f"Error: {e}")

print("\n=== Testing /v1/chat/completions with tools ===")
# Test 4: /v1/chat/completions with function calling
tools_payload = {
    "model": "gpt-3.5-turbo",
    "messages": [
        {"role": "user", "content": "What's the weather like?"}
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get weather information",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "location": {"type": "string"}
                    }
                }
            }
        }
    ]
}

try:
    resp = requests.post('https://api.openai.com/v1/chat/completions',
                        headers=headers,
                        json=tools_payload)
    print(f"Status: {resp.status_code}")
    print(f"Response: {resp.text[:500]}...")
except Exception as e:
    print(f"Error: {e}")

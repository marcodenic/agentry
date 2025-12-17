# Agentry Manual Test Plan

This document provides a step-by-step guide for manually verifying the Agentry system, focusing on the TUI and new features like Google Gemini support.

## 1. Setup

Ensure you have the necessary API keys set in your environment:
```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-..."
export GOOGLE_API_KEY="AIza..."
```

Build the latest version:
```bash
./build.sh
```

## 2. TUI Verification

### Basic UI Interaction
1.  **Launch:** Run `./agentry`. Verify the TUI opens with the welcome message.
2.  **Input:** Type "Hello, are you there?". Verify text appears in the input box.
3.  **Send:** Press Enter. Verify the message moves to the chat history.
4.  **Response:** Verify the agent responds (streaming text).
5.  **Scrolling:** If the history is long, verify you can scroll up/down (PageUp/PageDown or mouse wheel).

### View Management
1.  **Tabs:** Press `Tab` to switch between views (if multiple tabs are active).
2.  **Debug View:** Press `Ctrl+D`. Verify the Debug Log view opens.
3.  **Activity Feed:** Press `Ctrl+F`. Verify the Activity Feed sidebar toggles.
4.  **Resize:** Resize the terminal window. Verify the layout adjusts dynamically.

### Tool Usage (Visual)
1.  **Request:** Ask "List the files in the current directory."
2.  **Observation:** Watch for the "Tool Use" indicator or status update in the TUI.
3.  **Output:** Verify the file list is displayed in the chat.

## 3. Provider Verification

### Google (Gemini) Support
1.  **Configuration:**
    Create or edit `agentry.yaml` (or use flags) to specify the Google provider.
    ```yaml
    model:
      provider: google
      options:
        model: gemini-2.0-flash
    ```
2.  **Run:** `./agentry --config agentry.yaml` (or equivalent).
3.  **Verify:** Ask "What model are you?". Verify it identifies as Gemini (or at least responds).
4.  **Tool Use:** Ask "What is 25 * 48?". Verify it uses a tool (if calculator available) or computes it.
    *Note: Gemini tool use requires the model to support function calling.*

### OpenAI / Anthropic
1.  **Switch:** Change config to `openai` or `anthropic`.
2.  **Verify:** Run and ask a simple question to ensure no regression.

## 4. Edge Cases

1.  **Network Failure:** Disconnect internet and send a message. Verify graceful error handling (no crash).
2.  **Invalid Key:** Unset API key and run. Verify error message in TUI or exit.
3.  **Large Output:** Ask for a long story. Verify streaming continues smoothly.

## 5. Release Sign-off

- [ ] TUI compiles and runs.
- [ ] Basic chat works.
- [ ] Tool usage works.
- [ ] Google provider works.
- [ ] No panics/crashes during standard usage.

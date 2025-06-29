#!/usr/bin/env python3

"""
Interactive test for Agentry CLI chat mode with proper timing control.
This script sends commands one at a time to avoid LLM flooding.
"""

import subprocess
import time
import os
import sys
from threading import Thread
import signal

class AgentryTester:
    def __init__(self, workspace="/tmp/agentry-test-workspace"):
        self.workspace = workspace
        self.agentry_path = "/home/marco/Documents/GitHub/agentry/agentry.exe"
        self.process = None
        self.setup_workspace()
        
    def setup_workspace(self):
        """Create and change to test workspace"""
        os.makedirs(self.workspace, exist_ok=True)
        os.chdir(self.workspace)
        print(f"📁 Test workspace: {self.workspace}")
        
    def start_chat_session(self):
        """Start the agentry chat session"""
        print("🚀 Starting Agentry chat session...")
        self.process = subprocess.Popen(
            [self.agentry_path, "chat"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            universal_newlines=True
        )
        print("✅ Chat session started")
        
    def send_command(self, command, description, timeout=120):
        """Send a command and wait for processing"""
        print("\n" + "="*50)
        print(f"📤 Test: {description}")
        print(f"Command: {command}")
        print("="*50)
        
        if not self.process:
            print("❌ No active chat session")
            return
            
        try:
            # Send the command
            self.process.stdin.write(command + "\n")
            self.process.stdin.flush()
            print("✅ Command sent")
            
            # Wait a bit for processing
            print("⏳ Waiting for response...")
            time.sleep(10)  # Give LLM time to process
            
            # Check if process is still alive
            if self.process.poll() is None:
                print("🤖 Agent is still processing...")
                time.sleep(5)  # Additional wait
            
        except Exception as e:
            print(f"❌ Error sending command: {e}")
            
    def check_files(self):
        """Check what files were created"""
        print("\n📂 Current workspace contents:")
        try:
            files = os.listdir('.')
            if files:
                for f in files:
                    stat = os.stat(f)
                    print(f"  📄 {f} ({stat.st_size} bytes)")
                    if f.endswith('.txt') and stat.st_size < 1000:
                        try:
                            with open(f, 'r') as file:
                                content = file.read().strip()
                                print(f"     Content: {content}")
                        except:
                            pass
            else:
                print("  (empty)")
        except Exception as e:
            print(f"  Error listing files: {e}")
            
    def end_session(self):
        """End the chat session"""
        if self.process:
            print("\n🔚 Ending session...")
            try:
                self.process.stdin.write("/quit\n")
                self.process.stdin.flush()
                self.process.wait(timeout=10)
            except:
                self.process.terminate()
            self.process = None
            
    def run_tests(self):
        """Run the full test suite"""
        print("🧪 Agentry CLI Interactive Test (Python)")
        print("=========================================")
        
        self.start_chat_session()
        time.sleep(2)  # Let it initialize
        
        # Test sequence
        tests = [
            ("What are your capabilities as Agent 0? What tools do you have for team coordination?", 
             "Agent 0 self-awareness", 90),
            
            ("What is the current team status? Use the team_status tool to check.", 
             "Team status check", 60),
            
            ("Create a file called agent_test_file.txt with the content 'Hello from Agent 0'", 
             "File creation task", 90),
            
            ("/spawn coder \"Help with coding tasks\"", 
             "Spawn coder agent", 120),
            
            ("/list", 
             "List all agents", 30),
            
            ("What files are in the current directory? Please analyze them.", 
             "Workspace analysis", 60),
        ]
        
        for i, (command, description, timeout) in enumerate(tests, 1):
            print(f"\n🔄 Running test {i}/{len(tests)}")
            self.send_command(command, description, timeout)
            time.sleep(3)  # Brief pause between tests
            
            # Check files after file creation test
            if "Create a file" in command:
                time.sleep(2)
                self.check_files()
        
        # Final checks
        self.check_files()
        self.end_session()
        
        print("\n✅ All tests completed!")
        print("\n🔍 Summary of tests:")
        print("- Agent 0 capabilities and awareness")
        print("- Team status checking")
        print("- File creation")
        print("- Agent spawning")
        print("- Agent listing")
        print("- Workspace analysis")

def main():
    tester = AgentryTester()
    
    # Handle Ctrl+C gracefully
    def signal_handler(sig, frame):
        print("\n🛑 Test interrupted by user")
        tester.end_session()
        sys.exit(0)
        
    signal.signal(signal.SIGINT, signal_handler)
    
    try:
        tester.run_tests()
    except Exception as e:
        print(f"❌ Test failed with error: {e}")
        tester.end_session()
        sys.exit(1)

if __name__ == "__main__":
    main()

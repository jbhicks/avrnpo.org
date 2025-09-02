#!/bin/bash
# Launch Buffalo dev server and multitail log monitoring in tmux
SESSION=dev

# Kill any existing session
if tmux has-session -t $SESSION 2>/dev/null; then
  tmux kill-session -t $SESSION
fi

# Start new session with Buffalo server
tmux new-session -d -s $SESSION 'make dev'

# Split window and run multitail
tmux split-window -h -t $SESSION 'multitail logs/application.log logs/audit.log'

# Attach to the session
tmux attach -t $SESSION

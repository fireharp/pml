#!/bin/bash
WATCH_DIR="/Users/fireharp/Prog/Stuff/python-llm/impl1/results"
LOG_FILE="/tmp/kill_process_log.txt"

echo "Monitoring $WATCH_DIR for file changes..." | tee -a "$LOG_FILE"

while fswatch -1 "$WATCH_DIR"; do
  for file in "$WATCH_DIR"/*; do
    if [ ! -f "$file" ]; then
      continue  # Skip if it's not a file
    fi

    pids=$(lsof -t "$file")

    if [ -z "$pids" ]; then
      echo "$(date) - No processes found holding $file open" | tee -a "$LOG_FILE"
    else
      for pid in $pids; do
        if [ "$pid" != "$$" ]; then
          process_name=$(ps -p "$pid" -o comm=)
          echo "$(date) - Killing process $pid ($process_name) writing to $file" | tee -a "$LOG_FILE"

          kill -TERM "$pid" 2>/dev/null
          sleep 0.5

          if kill -0 "$pid" 2>/dev/null; then
            echo "$(date) - Process $pid ($process_name) did not exit, using SIGKILL" | tee -a "$LOG_FILE"
            kill -KILL "$pid" 2>/dev/null
          else
            echo "$(date) - Successfully terminated process $pid ($process_name)" | tee -a "$LOG_FILE"
          fi
        fi
      done
    fi
  done
done
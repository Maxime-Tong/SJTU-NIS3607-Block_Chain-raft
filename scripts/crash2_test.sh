#!/bin/bash
go build -o ../cmd/node ../cmd/node.go

tmux new -d -s node0 "../cmd/node -i 0 -t 30"
tmux new -d -s node1 "../cmd/node -i 1 -t 30"
tmux new -d -s node2 "../cmd/node -i 2 -t 30"
tmux new -d -s node3 "../cmd/node -i 3 -t 30"
tmux new -d -s node4 "../cmd/node -i 4 -t 30"
tmux new -d -s node5 "../cmd/node -i 5 -t 30"
tmux new -d -s node6 "../cmd/node -i 6 -t 30"

sleep 10
tmux kill-window -t node$1

sleep 5
tmux kill-window -t node$2

sleep 25

python3 check.py 7 $1 $2

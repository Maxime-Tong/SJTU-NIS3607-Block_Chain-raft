go build -o ../cmd/node ../cmd/node.go

../cmd/node -i 0 -t $1 &
../cmd/node -i 1 -t $1 &
../cmd/node -i 2 -t $1 &
../cmd/node -i 3 -t $1 &
../cmd/node -i 4 -t $1 &
../cmd/node -i 5 -t $1 &
../cmd/node -i 6 -t $1 &

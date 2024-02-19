set-env:
	go run cmd/env/main.go --set $(NAME) $(VALUE)

get-env:
	go run cmd/env/main.go printenv $(NAME)

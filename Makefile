
NAME = tazama

all : set $(NAME)

set:
	go mod tidy

$(NAME):
	go build -o bin/

clean:
	rm $(NAME)


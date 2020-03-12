PROJECT_NAME = heis

.PHONY: clean

build :
	go build -o $(PROJECT_NAME) main.go

clean :
	rm -rf $(PROJECT_NAME) 2> /dev/null

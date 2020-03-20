PROJECT_NAME = heis

.PHONY: clean
.PHONY: run1
.PHONY: run2

build :
	go build -o $(PROJECT_NAME) main.go

run1 :
	./heis | tee out1.log

run2 :
	./heis -port 15658 | tee out2.log

clean :
	rm -rf $(PROJECT_NAME) 2> /dev/null

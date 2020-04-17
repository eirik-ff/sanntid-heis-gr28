PROJECT_NAME = heis
LOGS_DIR = ./logs

.PHONY: clean
.PHONY: run1
.PHONY: run2
.PHONY: run3
.PHONY: packetloss
.PHONY: packetlossoff

FROMFILE = 

build :
	go build -o $(PROJECT_NAME) main.go

logs/ :
	mkdir $(LOGS_DIR)

run1 : logs/
	./heis --port 15657 --wd 57005 $(FROMFILE) | tee logs/out15657.log

run2 : logs/
	./heis --port 15658 --wd 57006 $(FROMFILE) | tee logs/out15658.log

run3 : logs/
	./heis --port 15659 --wd 57007 $(FROMFILE) | tee logs/out15659.log
    
packetloss :
	sudo iptables -A INPUT -p tcp --dport 15657 -j ACCEPT
	sudo iptables -A INPUT -p tcp --sport 15657 -j ACCEPT
	sudo iptables -A INPUT -m statistic --mode random --probability 0.2 -j DROP

packetlossoff :
	sudo iptables -F

clean :
	rm -rf $(PROJECT_NAME) 2> /dev/null
	rm -rf $(LOGS_DIR) 2> /dev/null
	rm -rf *.log 2> /dev/null


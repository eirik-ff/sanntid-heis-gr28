PROJECT_NAME = heis

.PHONY: clean
.PHONY: run1
.PHONY: run2
.PHONY: run3
.PHONY: packetloss
.PHONY: packetlossoff

FROMFILE = false

build :
	go build -o $(PROJECT_NAME) main.go

run1 :
	./heis -port 15657 -wd 57005 -fromfile $(FROMFILE) | tee out15657.log

run2 :
	./heis -port 15658 -wd 57006 -fromfile $(FROMFILE) | tee out15658.log

run3 :
	./heis -port 15659 -wd 57007 -fromfile $(FROMFILE) | tee out15659.log
    
packetloss :
	sudo iptables -A INPUT -p tcp --dport 15657 -j ACCEPT
	sudo iptables -A INPUT -p tcp --sport 15657 -j ACCEPT
	sudo iptables -A INPUT -m statistic --mode random --probability 0.2 -j DROP

packetlossoff :
	sudo iptables -F

clean :
	rm -rf $(PROJECT_NAME) 2> /dev/null

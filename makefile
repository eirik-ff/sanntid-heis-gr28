PROJECT_NAME = heis
LOGS_DIR = ./logs
WD_SUBMOD_DIR = ./watchdog-go-submod
CWD = $(shell pwd)
WD_MSG = "28-IAmAlive"

.PHONY: run1
.PHONY: run2
.PHONY: run3
.PHONY: start1
.PHONY: start2
.PHONY: start3

.PHONY: help
.PHONY: clean
.PHONY: packetloss
.PHONY: packetlossoff

FROMFILE = 


#####################
### Build targets ###
#####################
build : logs/
	go build -o $(PROJECT_NAME) main.go

buildall : build
	cd $(WD_SUBMOD_DIR) && make
	cp $(WD_SUBMOD_DIR)/wd ./wd

logs/ :
	mkdir $(LOGS_DIR)

#####################
### Start targets ###
#####################
start1 : logs/
	./wd --port=57005 --message=$(WD_MSG) --exec='$(CWD)/heis --port=15657 --wd=57005 --fromfile'

start2 : logs/
	./wd --port=57006 --message=$(WD_MSG) --exec='$(CWD)/heis --port=15658 --wd=57006 --fromfile'

start3 : logs/
	./wd --port=57007 --message=$(WD_MSG) --exec='$(CWD)/heis --port=15659 --wd=57007 --fromfile'

###################
### Run targets ###
###################
run1 : logs/
	./heis --port 15657 --wd 57005 $(FROMFILE) | tee $(LOG_DIR)/out15657.log

run2 : logs/
	./heis --port 15658 --wd 57006 $(FROMFILE) | tee $(LOG_DIR)/out15658.log

run3 : logs/
	./heis --port 15659 --wd 57007 $(FROMFILE) | tee $(LOG_DIR)/out15659.log
    
#####################
### Other targets ###
#####################
help :
	@echo Targets:
	@echo '  build:    compiles only elevator program.'
	@echo '  buildall: compiles elevator program and watchdog.'
	@echo '  startN:   starts the watchdog which in turn starts the elevator N.'
	@echo '  runN:     starts the elevator N.'
	@echo ''
	@echo 'cwd: $(CWD)'
	@echo 'watchdog msg: $(WD_MSG)'

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
	cd $(WD_SUBMOD_DIR) && make clean


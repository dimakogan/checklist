
all: log.pdf sim.pdf

log.pdf: plotLog.py
	python plotLog.py < sbtunnel.log

sim.pdf: plotSim.py
	python plotSim.py < sbtunnel.log

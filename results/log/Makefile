
all: log.pdf sim.pdf trace_client.pdf

log.pdf: plotLog.py
	python plotLog.py < sbtunnel.log

sim.pdf: plotSim.py
	python plotSim.py < sbtunnel.log

trace_client.pdf: plotTrace.py
	python plotTrace.py boosted.txt dpf.txt matrix.txt

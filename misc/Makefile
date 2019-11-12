CXX = g++
CXXFLAGS = -O3 -g -fPIC -pthread -std=c++0x

default: simpleperf

simpleperf.o: simpleperf.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

simpleperf: simpleperf.o
	$(CXX) $(CXXFLAGS) $^ -o $@

clean:
	rm -rf *.o simpleperf
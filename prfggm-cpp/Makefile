all: main

main: main.cpp pset_ggm.h pset_ggm.cpp Defines.cpp AES.cpp
	g++ -msse2 -msse -march=native -maes -Ofast -o main main.cpp pset_ggm.cpp Defines.cpp AES.cpp

clean: 
	rm main

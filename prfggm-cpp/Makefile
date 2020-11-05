all: psetggm.lib main


HEADERS := pset_ggm.h
SRC :=  pset_ggm.cpp Defines.cpp AES.cpp

OBJ := $(addsuffix .o,$(basename $(SRC)))

FLAGS := -msse2 -msse -march=native -maes -Ofast

main: $(HEADERS) $(SRC)
	g++ $(FLAGS) -o main main.cpp $(SRC)

%.o: %.cpp $(HEADERS)
	g++ $(FLAGS) -c $< -o $@

libpsetggm.a: $(OBJ)
	echo $(FLAGS)
	ar rcs $@ $(OBJ)

clean: 
	rm main

all: $(LIB) main

LIB := libpsetggm.a
HEADERS := pset_ggm.h
SRC :=  pset_ggm.cpp Defines.cpp AES.cpp

OBJ := $(addsuffix .o,$(basename $(SRC)))

FLAGS := -msse2 -msse -march=native -maes -g # -Ofast 

main: $(HEADERS) $(SRC) main.cpp
	g++ $(FLAGS) -o main main.cpp $(SRC)

%.o: %.cpp $(HEADERS)
	g++ $(FLAGS) -c $< -o $@

$(LIB): $(OBJ)
	echo $(FLAGS)
	ar rcs $@ $(OBJ)

clean: 
	rm main

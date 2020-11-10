
SRC_DIR := ..
LIB := libpsetggm.a
HEADERS := $(addprefix $(SRC_DIR)/, pset_ggm.h)
SRC := $(addprefix $(SRC_DIR)/, pset_ggm.cpp Defines.cpp AES.cpp)

OBJ := $(addsuffix .o,$(basename $(SRC)))

FLAGS := -msse2 -msse -march=native -maes  -g #-Ofast

all: $(LIB) main

main: $(HEADERS) $(SRC) main.cpp
	g++ $(FLAGS) -o main main.cpp $(SRC)

%.o: %.cpp $(HEADERS)
	g++ $(FLAGS) -c $< -o $@

$(LIB): $(OBJ)
	echo $(FLAGS)
	ar rcs $@ $(OBJ)

clean: 
	rm main $(LIB)

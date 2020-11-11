
SRC_DIR := ..
LIB := libpsetggm.a
HEADERS := $(addprefix $(SRC_DIR)/, pset_ggm.h)
SRC := $(addprefix $(SRC_DIR)/, pset_ggm.cpp Defines.cpp AES.cpp xor.cpp)

OBJ := $(addsuffix .o,$(basename $(SRC)))

FLAGS := -msse2 -msse -march=native -maes -Ofast

all: $(LIB) main

main: $(HEADERS) $(SRC) main.cpp
	g++ $(FLAGS) -I.. -o main main.cpp $(SRC)

distinct: $(HEADERS) $(SRC) distinct.cpp
	g++ $(FLAGS) -I.. -o distinct distinct.cpp $(SRC)

test_xor: $(HEADERS) $(SRC) test_xor.cpp
	g++ $(FLAGS) -I.. -o test_xor test_xor.cpp $(SRC)


%.o: %.cpp $(HEADERS)
	g++ $(FLAGS) -c $< -o $@

$(LIB): $(OBJ)
	echo $(FLAGS)
	ar rcs $@ $(OBJ)

clean: 
	rm main $(LIB)

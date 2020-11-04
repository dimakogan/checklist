#include "pset_ggm.h"

#include <chrono> 
#include <iostream>

using namespace std::chrono; 

int main(int argc, char** argv) {
    enum ARGS {
        PROGRAM_NAME = 0,
        SET_SIZE,
        NUM_ARGS
    };

    if (argc < NUM_ARGS) {
        printf("Usage: %s <NUM_ARGS>\n", argv[PROGRAM_NAME]);
        return 1;
    }

    uint32_t set_size = atoi(argv[SET_SIZE]);

    SetGenerator gen(20000, set_size);
    __m128i seed = _mm_setr_epi32(0, 0, 0, 0);
    std::vector<uint32_t> out(set_size);
    
    auto start = high_resolution_clock::now(); 
    for (int i=0; i < 1000; ++i) {
        gen.Eval(seed, out.data());
    }
    auto stop = high_resolution_clock::now(); 


    std::cout   << "Eval time: " 
                << duration_cast<microseconds>(stop - start).count()/1000
                << std::endl;
    // for (const auto& i : out) {
    //     std::cout << i << " ";
    // }
    // std::cout<< std::endl;

    return 0;
}
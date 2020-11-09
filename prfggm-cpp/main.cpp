#include "pset_ggm.h"

#include <chrono> 
#include <cstdint>
#include <iostream>
#include <vector>

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

    const uint8_t seed[] = {0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0};
    std::vector<uint32_t> out(set_size*2);
    
    auto gen = init_generator(1024*1024, set_size);
    auto start = high_resolution_clock::now(); 
    for (int i=0; i < 100000; ++i) {
        pset_ggm_eval(gen, seed, out.data());
    }
    auto stop = high_resolution_clock::now(); 

    std::cout   << "Eval time: " 
                << duration_cast<microseconds>(stop - start).count()/100000
                << std::endl;
    // for (int i =0; i < set_size; ++i) {
    //     std::cout << out[i] << " ";
    // }
    // std::cout<< std::endl;


    std::vector<uint32_t> punctured(set_size*2);
    for (int pos = 0; pos < set_size; ++pos) {
        unsigned int pset_size; 
        auto pset = pset_ggm_punc(gen, seed, pos, &pset_size);
        // for (int i=0; i < pset_size; i++) {
        //     printf("%02x", pset[i]);
        // }
        pset_ggm_eval_punc(gen, pset, pos, punctured.data());
        for (int i =0; i < set_size; ++i) {
            if ((i != pos) && (out[i] != punctured[i])) {
                std::cout << "Differ: " << out[i] << ", " << punctured[i];
            }
        }
        free_pset(pset);
    } 
    free_generator(gen);

    return 0;
}
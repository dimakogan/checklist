#include "pset_ggm.h"

#include <chrono> 
#include <cstdint>
#include <iostream>
#include <unordered_set>
#include <vector>

using namespace std::chrono; 

int main(int argc, char** argv) {
    enum ARGS {
        PROGRAM_NAME = 0,
        SET_SIZE,
        NUM_ARGS
    };

    if (argc < NUM_ARGS) {
        printf("Usage: %s <SET_SIZE>\n", argv[PROGRAM_NAME]);
        return 1;
    }

    uint32_t univ_size = 2 * 1024*1024;
    uint32_t set_size = atoi(argv[SET_SIZE]);

    const uint8_t seed[] = {0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0};

    unsigned int worksize = workspace_size(univ_size, set_size);
    std::vector<uint8_t> workspace(worksize);

    generator* gen = pset_ggm_init(univ_size, set_size, workspace.data());

    std::vector<long long unsigned int> out(set_size);
    auto start = high_resolution_clock::now(); 
    for (int i=0; i < 100000; ++i) {
        *(int*)(seed) = i;
        pset_ggm_eval(gen, seed, out.data());
    }
    auto stop = high_resolution_clock::now(); 

    std::cout   << "Eval time: " 
                << duration_cast<nanoseconds>(stop - start).count()/100000
                << "ns"
                << std::endl;
    // for (int i =0; i < set_size; ++i) {
    //     std::cout << out[i] << " ";
    // }
    // std::cout<< std::endl;

    
    unsigned int pset_size = pset_buffer_size(gen);
    std::vector<uint8_t> pset(pset_size);

    std::vector<long long unsigned int> pelems(set_size);
    for (int pos = 0; pos < set_size; ++pos) {
        pset_ggm_punc(gen, seed, pos, pset.data());
        pset_ggm_eval_punc(gen, pset.data(), pos, pelems.data());
        int compare_pos = 0;
        for (int i =0; i < set_size; ++i) {
            if (i == pos) {
                continue;
            }
            //std::cout  << pelems[i] << ", ";
            if ((i != pos) && (out[i] != pelems[compare_pos])) {
                std::cout << "Differ: " << out[i] << ", " << pelems[compare_pos];
            }
            compare_pos++;
        }
        //std::cout << std::endl;
    } 

    return 0;
}
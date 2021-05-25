#include <algorithm>
#include <chrono>
#include <iostream>
#include <pset_ggm.h>
#include <unordered_set>
#include <vector>

#include "flat_hash_map.hpp"

using namespace std::chrono;

const uint32_t univ_size = 2 * 1024 * 1024;
const uint32_t set_size = 1414;
std::vector<uint32_t> present(univ_size);
uint32_t present_mark = 0;

bool distinct_lin_space(int univ_size, const long long unsigned int *elems, unsigned int num_elems)
{
    present_mark++;
    for (int i = 0; i < num_elems; ++i)
    {
        uint32_t elem = elems[i];
        uint32_t *ptr = &present[elem];
        if (*ptr == present_mark)  {
            return false;
        }
        *ptr = present_mark;
    }
    return true;
}

ska::flat_hash_set<long long unsigned int> present_hash;

bool distinct_hash(int univ_size, const long long unsigned int *elems, unsigned int num_elems)
{
    present_hash.clear();
    present_hash.reserve(num_elems);
    for (int i = 0; i < num_elems; ++i)
    {
        auto elem = elems[i];
        present_hash.insert(elem);
    }
    return (present_hash.size() == num_elems);
}

int main(int argc, char **argv)
{
    const uint8_t seed[] = {0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0};

    
    unsigned int worksize = workspace_size(univ_size, set_size);
    std::vector<uint8_t> workspace(worksize);

    generator *gen = pset_ggm_init(univ_size, set_size, workspace.data());

    std::vector<long long unsigned int> elems(set_size);
    pset_ggm_eval(gen, seed, elems.data());


    const int num_iterations = 100000;
    auto start = high_resolution_clock::now();
    bool ok = true;
    for (int j = 0; j < num_iterations; ++j)
    {
        *(int*)(seed) = j;
        pset_ggm_eval(gen, seed, elems.data());
        ok = distinct_lin_space(univ_size, elems.data(), elems.size());
    }
    auto stop = high_resolution_clock::now();
    std::cout << "Distinct linear space: " << ok << ", time: " << duration_cast<nanoseconds>(stop - start).count() / num_iterations << "ns" << std::endl;

    present_hash.reserve(elems.size());

    start = high_resolution_clock::now();
    ok = true;
    for (int j = 0; j < num_iterations; ++j)
    {
        ok = distinct_hash(univ_size, elems.data(), elems.size());
    }
    stop = high_resolution_clock::now();
    std::cout << "Distinct hash: " << ok << ", time: " << duration_cast<nanoseconds>(stop - start).count() / num_iterations << "ns" << std::endl;


    int num_differ = 0;
    int num_distinct= 0;
    int num_collisions = 0;

    for (int j = 0; j < num_iterations; ++j)
    {
        *(int*)(seed) = j;
        pset_ggm_eval(gen, seed, elems.data());
        if (distinct_lin_space(univ_size, elems.data(), elems.size())) {
            num_distinct++;
        }
        int collisions = 0;
        if ((distinct(gen, elems.data(), elems.size()) != false) != distinct_lin_space(univ_size, elems.data(), elems.size())) {
            num_differ++;
            // std::cout << "Different result: " << j  
            //     << " custom: " << distinct_custom(univ_size, elems) 
            //     << ", lin_space: " << distinct_lin_space(univ_size, elems.data(), elems.size()) 
            //     << std::endl;
            // for ( auto e : elems) {
            //     std::cout << e << " ";
            // }
            // std::cout << std::endl;
        }
        num_collisions+= collisions;
    }
    std::cout << "Disagreements between lin_space and custom: " << num_differ << " out of " << num_iterations << " (" << num_distinct << " distinct)" << std::endl;

    std::cout << "Average number of collisions: " << num_collisions/num_iterations << std::endl;

    start = high_resolution_clock::now();
    ok = true;
    int collisions;
    for (int j = 0; j < num_iterations; ++j)
    {
        *(int*)(seed) = j;
        pset_ggm_eval(gen, seed, elems.data());
        ok = (distinct(gen, elems.data(), elems.size()) != false);
    }
    stop = high_resolution_clock::now();
    std::cout << "Distinct custom: " << ok << ", time: " << duration_cast<nanoseconds>(stop - start).count() / num_iterations << "ns" << std::endl;

    return 0;
}
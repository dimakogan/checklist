#include <chrono>
#include <iostream>
#include <pset_ggm.h>
#include <unordered_set>
#include <vector>

using namespace std::chrono;

uint32_t univ_size = 1024 * 1024;
uint32_t set_size = 1000;
std::vector<uint32_t> present(univ_size);
uint32_t present_mark = 0;

bool distinct_lin_space(int univ_size, const long long unsigned int *elems, unsigned int num_elems)
{
    bool ok = true;

    present_mark++;
    for (int i = 0; i < num_elems; ++i)
    {
        uint32_t elem = elems[i];
        uint32_t *ptr = &present[elem];
        ok &= ~(*ptr == present_mark);
        *ptr = present_mark;
    }
    return ok;
}

std::unordered_set<long long unsigned int> present_hash;

bool distinct_hash(int univ_size, const long long unsigned int *elems, unsigned int num_elems)
{
    present_hash.clear();
    bool ok = true;
    for (int i = 0; i < num_elems; ++i)
    {
        auto elem = elems[i];
        if (present_hash.find(elem) != present_hash.end())
        {
            ok = false;
        }
        else
        {
            present_hash.insert(elem);
        }
    }
    return ok;
}

int main(int argc, char **argv)
{
    const uint8_t seed[] = {0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0};

    unsigned int worksize = workspace_size(univ_size, set_size);
    std::vector<uint8_t> workspace(worksize);

    generator *gen = pset_ggm_init(univ_size, set_size, workspace.data());

    std::vector<long long unsigned int> elems(set_size);
    pset_ggm_eval(gen, seed, elems.data());

    auto start = high_resolution_clock::now();
    bool ok = true;
    for (int j = 0; j < 1000; ++j)
    {
        ok = distinct_lin_space(univ_size, elems.data(), elems.size());
    }
    auto stop = high_resolution_clock::now();
    std::cout << "Distinct linear space: " << ok << ", time: " << duration_cast<nanoseconds>(stop - start).count() / 1000 << "ns" << std::endl;

    present_hash.reserve(elems.size());

    start = high_resolution_clock::now();
    ok = true;
    for (int j = 0; j < 1000; ++j)
    {
        ok = distinct_hash(univ_size, elems.data(), elems.size());
    }
    stop = high_resolution_clock::now();
    std::cout << "Distinct hash: " << ok << ", time: " << duration_cast<nanoseconds>(stop - start).count() / 1000 << "ns" << std::endl;

    return 0;
}
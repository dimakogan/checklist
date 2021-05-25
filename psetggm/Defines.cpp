#include "Defines.h"
#include <sstream>
#include <cstring>


const block ZeroBlock = _mm_set_epi64x(0, 0);
const block LSBBlock = _mm_set_epi64x(0, 1);
const block MSBBlock = _mm_set_epi64x(1ULL<<63, 0);
const block AllOneBlock = _mm_set_epi64x(uint64_t(-1), uint64_t(-1));
const block TestBlock = ([]() {block aa; memset(&aa, 0xaa, sizeof(block)); return aa; })();

void split(const std::string &s, char delim, std::vector<std::string> &elems) {
    std::stringstream ss(s);
    std::string item;
    while (std::getline(ss, item, delim)) {
        elems.push_back(item);
    }
}

std::vector<std::string> split(const std::string &s, char delim) {
    std::vector<std::string> elems;
    split(s, delim, elems);
    return elems;
}

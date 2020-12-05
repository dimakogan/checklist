#include "answer.h"

#include "pset_ggm.h"
#include "xor.h"

#include <vector>

extern "C" {

void answer(const uint8_t* pset, unsigned int pos, unsigned int univ_size, unsigned int set_size, unsigned int shift,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out) {
    auto worksize = workspace_size(univ_size, set_size+1);
    auto workspace = (uint8_t*)malloc(worksize+set_size*sizeof(long long unsigned int));
    auto gen = pset_ggm_init(univ_size, set_size+1, workspace);

    auto elems = (long long unsigned int*)(workspace+worksize);
    pset_ggm_eval_punc(gen, pset, pos, elems);

    for (int i = 0; i < set_size; i++) 
        elems[i] = ((elems[i]+shift)%univ_size)*row_len;


    xor_rows(db, db_len, elems, set_size, block_len, out);
    free(workspace);
}

} // extern "C"
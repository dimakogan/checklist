#include <stdint.h>
#include <immintrin.h>
#include <smmintrin.h>

#ifdef __cplusplus
extern "C" {
#endif

struct generator;

generator* init_generator(unsigned int univ_size, unsigned int set_size);
void pset_ggm_eval(generator* gen, const uint8_t* seed, uint32_t* out);
uint8_t* pset_ggm_punc(generator* gen, const uint8_t* seed, unsigned int pos, unsigned int* out_size);
void pset_ggm_eval_punc(generator* gen, const uint8_t* pset, unsigned int pos, uint32_t* out);

void free_pset(uint8_t* pset);
void free_generator(generator* gen);

#ifdef __cplusplus
}
#endif
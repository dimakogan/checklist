#include <stdint.h>
#include <immintrin.h>
#include <smmintrin.h>

#ifdef __cplusplus
extern "C" {
#endif

void pset_ggm_eval(unsigned int	 univ_size, unsigned int set_size, const unsigned char* seed, unsigned long* out);


#ifdef __cplusplus
}
#endif
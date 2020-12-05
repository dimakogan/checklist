#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Fast combined Answer function to save on allocations.
void answer(const uint8_t* pset, unsigned int pos, unsigned int univ_size, unsigned int set_size, unsigned int shift,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out);


#ifdef __cplusplus
} // extern "C" 
#endif


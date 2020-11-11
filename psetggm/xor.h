#ifdef __cplusplus
extern "C" {
#endif



void xor_rows(const uint8_t* db, unsigned int row_len, unsigned int num_rows, 
    const long long unsigned int* elems, unsigned int num_elems, uint8_t* out);


#ifdef __cplusplus
}
#endif

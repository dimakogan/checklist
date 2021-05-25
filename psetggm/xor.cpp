#include <cstdint>
#include <cstdio>
#include <cstring>
#include "intrinsics.h"

extern "C"
{
#ifdef __AVX2__

    void xor_rows(const uint8_t* db, unsigned int db_len, 
        const long long unsigned int* elems, unsigned int num_elems, 
        unsigned int block_len, uint8_t* out)
    {
        memset(out, 0, block_len);
        for (int i = 0; i < num_elems; i++)
        {
            if (elems[i] > (db_len-block_len))
            {
                continue;
            }
            __m256i *block = (__m256i *)(db + elems[i]);
            for (int b = 0; b < (block_len / 32); b++)
            {
                __m256i out256 = _mm256_loadu_si256((__m256i *)out + b);
                __m256i elem = _mm256_loadu_si256(block + b);
                out256 = _mm256_xor_si256(out256, elem);
                _mm256_storeu_si256((__m256i *)out + b,  out256);
            }
        }
        if ((block_len % 32) == 0)
            return;

        bool use128 = ((block_len % 32) >= 16);
        bool use64 = ((block_len % 16) >= 8);
        bool use32 = ((block_len % 8) >= 4);
        bool use16 = ((block_len % 4) >= 2);
        bool use8 = ((block_len % 2) >= 1);

        size_t off128 = block_len - (block_len % 32);
        size_t off64 = block_len - (block_len % 16);
        size_t off32 = block_len - (block_len % 8);
        size_t off16 = block_len - (block_len % 4);
        size_t off8 = block_len - 1;

        __m128i out128 = _mm_setzero_si128();
        uint64_t out64 = 0;
        uint32_t out32 = 0;
        uint16_t out16 = 0;
        uint8_t out8 = 0;
        for (int i = 0; i < num_elems; i++)
        {
            if (elems[i] > db_len)
            {
                continue;
            }
            const uint8_t *block = db + elems[i];
            if (use128)
            {
                __m128i elem = _mm_load_si128((__m128i *)(block + off128));
                out128 = _mm_xor_si128(out128, elem);
            }
            if (use64)
            {
                out64 ^= *(uint64_t *)(block + off64);
            }
            if (use32)
            {
                out32 ^= *(uint32_t *)(block + off32);
            }
            if (use16)
            {
                out16 ^= *(uint16_t *)(block + off16);
            }
            if (use8)
            {
                out8 ^= *(uint8_t *)(block + off8);
            }
        }
        if (use128)
        {
            _mm_storeu_si128((__m128i *)(out + off128), out128);
        }
        if (use64)
        {
            *(uint64_t *)(out + off64) = out64;
        }
        if (use32)
        {
            *(uint32_t *)(out + off32) = out32;
        }
        if (use16)
        {
            *(uint16_t *)(out + off16) = out16;
        }
        if (use8)
        {
            *(uint8_t *)(out + off8) = out8;
        }
    }

    // Copied from:  https://github.com/dkales/dpf-cpp/blob/master/hashdatastore.cpp
    void xor_hashes_by_bit_vector(const uint8_t* db, unsigned int db_len, 
        const uint8_t* indexing, uint8_t* out) {
    // Optimize for a hash list of 32-bytes hashes.
        __m256i result = _mm256_set_epi64x(0,0,0,0);
        __m256i results[8][2] = {
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result}, };
        
        int len_full_bytes = (db_len/32)&(~0x7);
        for(size_t i = 0; i < len_full_bytes; i+=8) {
            uint8_t tmp = indexing[i/8];
            results[0][(tmp>>0)&1] = _mm256_xor_si256(results[0][1], (((__m256i*)db))[i]);
            results[1][(tmp>>1)&1] = _mm256_xor_si256(results[1][1], (((__m256i*)db))[i+1]);
            results[2][(tmp>>2)&1] = _mm256_xor_si256(results[2][1], (((__m256i*)db))[i+2]);
            results[3][(tmp>>3)&1] = _mm256_xor_si256(results[3][1], (((__m256i*)db))[i+3]);
            results[4][(tmp>>4)&1] = _mm256_xor_si256(results[4][1], (((__m256i*)db))[i+4]);
            results[5][(tmp>>5)&1] = _mm256_xor_si256(results[5][1], (((__m256i*)db))[i+5]);
            results[6][(tmp>>6)&1] = _mm256_xor_si256(results[6][1], (((__m256i*)db))[i+6]);
            results[7][(tmp>>7)&1] = _mm256_xor_si256(results[7][1], (((__m256i*)db))[i+7]);
        }
        for (size_t i = len_full_bytes; i < db_len/32; i++) {
            uint8_t tmp = indexing[i/8];
            results[i%8][(tmp>>(i%8))&1] = _mm256_xor_si256(results[i%8][1], (((__m256i*)db))[i]);
        }
        result = _mm256_xor_si256(results[0][1], results[1][1]);
        result = _mm256_xor_si256(result, results[2][1]);
        result = _mm256_xor_si256(result, results[3][1]);
        result = _mm256_xor_si256(result, results[4][1]);
        result = _mm256_xor_si256(result, results[5][1]);
        result = _mm256_xor_si256(result, results[6][1]);
        result = _mm256_xor_si256(result, results[7][1]);
        _mm256_storeu_si256((__m256i *)out, result);
        return;
}

#else

    void xor_rows(const uint8_t* db, unsigned int db_len, 
        const long long unsigned int* elems, unsigned int num_elems, 
        unsigned int block_len, uint8_t* out)
    {
        memset(out, 0, block_len);
        for (int i = 0; i < num_elems; i++)
        {
            if (elems[i] > (db_len-block_len))
            {
                continue;
            }
            __m128i *block = (__m128i *)(db + elems[i]);
            for (int b = 0; b < (block_len / 16); b++)
            {
                __m128i out128 = _mm_loadu_si128((__m128i *)out + b);
                __m128i elem = _mm_loadu_si128(block + b);
                out128 = _mm_xor_si128(out128, elem);
                _mm_storeu_si128((__m128i *)out + b,  out128);
            }
        }
        if ((block_len % 16) == 0)
            return;

        bool use64 = ((block_len % 16) >= 8);
        bool use32 = ((block_len % 8) >= 4);
        bool use16 = ((block_len % 4) >= 2);
        bool use8 = ((block_len % 2) >= 1);

        size_t off64 = block_len - (block_len % 16);
        size_t off32 = block_len - (block_len % 8);
        size_t off16 = block_len - (block_len % 4);
        size_t off8 = block_len - 1;

        uint64_t out64 = 0;
        uint32_t out32 = 0;
        uint16_t out16 = 0;
        uint8_t out8 = 0;
        for (int i = 0; i < num_elems; i++)
        {
            if (elems[i] > db_len)
            {
                continue;
            }
            const uint8_t *block = db + elems[i];
            if (use64)
            {
                out64 ^= *(uint64_t *)(block + off64);
            }
            if (use32)
            {
                out32 ^= *(uint32_t *)(block + off32);
            }
            if (use16)
            {
                out16 ^= *(uint16_t *)(block + off16);
            }
            if (use8)
            {
                out8 ^= *(uint8_t *)(block + off8);
            }
        }
        if (use64)
        {
            *(uint64_t *)(out + off64) = out64;
        }
        if (use32)
        {
            *(uint32_t *)(out + off32) = out32;
        }
        if (use16)
        {
            *(uint16_t *)(out + off16) = out16;
        }
        if (use8)
        {
            *(uint8_t *)(out + off8) = out8;
        }
    }

    // Copied from:  https://github.com/dkales/dpf-cpp/blob/master/hashdatastore.cpp
    void xor_hashes_by_bit_vector(const uint8_t* db, unsigned int db_len, 
        const uint8_t* indexing, uint8_t* out) {
    // Optimize for a hash list of 32-bytes hashes.
        __m128i result = _mm_set_epi64x(0,0);
        __m128i results_hi[8][2] = {
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result}, };
        __m128i results_lo[8][2] = {
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result},
                {result, result}, };


        int len_full_bytes = (db_len/32)&(~0x7);
        for(size_t i = 0; i < len_full_bytes; i+=8) {
            uint8_t tmp = indexing[i/8];
            results_lo[0][(tmp>>0)&1] = _mm_xor_si128(results_lo[0][1], (((__m128i*)db))[2*i]);
            results_lo[1][(tmp>>1)&1] = _mm_xor_si128(results_lo[1][1], (((__m128i*)db))[2*i+2]);
            results_lo[2][(tmp>>2)&1] = _mm_xor_si128(results_lo[2][1], (((__m128i*)db))[2*i+4]);
            results_lo[3][(tmp>>3)&1] = _mm_xor_si128(results_lo[3][1], (((__m128i*)db))[2*i+6]);
            results_lo[4][(tmp>>4)&1] = _mm_xor_si128(results_lo[4][1], (((__m128i*)db))[2*i+8]);
            results_lo[5][(tmp>>5)&1] = _mm_xor_si128(results_lo[5][1], (((__m128i*)db))[2*i+10]);
            results_lo[6][(tmp>>6)&1] = _mm_xor_si128(results_lo[6][1], (((__m128i*)db))[2*i+12]);
            results_lo[7][(tmp>>7)&1] = _mm_xor_si128(results_lo[7][1], (((__m128i*)db))[2*i+14]);
            results_hi[0][(tmp>>0)&1] = _mm_xor_si128(results_hi[0][1], (((__m128i*)db))[2*i+1]);
            results_hi[1][(tmp>>1)&1] = _mm_xor_si128(results_hi[1][1], (((__m128i*)db))[2*i+3]);
            results_hi[2][(tmp>>2)&1] = _mm_xor_si128(results_hi[2][1], (((__m128i*)db))[2*i+5]);
            results_hi[3][(tmp>>3)&1] = _mm_xor_si128(results_hi[3][1], (((__m128i*)db))[2*i+7]);
            results_hi[4][(tmp>>4)&1] = _mm_xor_si128(results_hi[4][1], (((__m128i*)db))[2*i+9]);
            results_hi[5][(tmp>>5)&1] = _mm_xor_si128(results_hi[5][1], (((__m128i*)db))[2*i+11]);
            results_hi[6][(tmp>>6)&1] = _mm_xor_si128(results_hi[6][1], (((__m128i*)db))[2*i+13]);
            results_hi[7][(tmp>>7)&1] = _mm_xor_si128(results_hi[7][1], (((__m128i*)db))[2*i+15]);
        }
        for (size_t i = len_full_bytes; i < db_len/32; i++) {
            uint8_t tmp = indexing[i/8];
            results_lo[i%8][(tmp>>(i%8))&1] = _mm_xor_si128(results_lo[i%8][1], (((__m128i*)db))[2*i]);
            results_hi[i%8][(tmp>>(i%8))&1] = _mm_xor_si128(results_hi[i%8][1], (((__m128i*)db))[2*i+1]);
        }
        result = _mm_xor_si128(results_lo[0][1], results_lo[1][1]);
        result = _mm_xor_si128(result, results_lo[2][1]);
        result = _mm_xor_si128(result, results_lo[3][1]);
        result = _mm_xor_si128(result, results_lo[4][1]);
        result = _mm_xor_si128(result, results_lo[5][1]);
        result = _mm_xor_si128(result, results_lo[6][1]);
        result = _mm_xor_si128(result, results_lo[7][1]);
        _mm_storeu_si128((__m128i *)out, result);
        result = _mm_xor_si128(results_hi[0][1], results_hi[1][1]);
        result = _mm_xor_si128(result, results_hi[2][1]);
        result = _mm_xor_si128(result, results_hi[3][1]);
        result = _mm_xor_si128(result, results_hi[4][1]);
        result = _mm_xor_si128(result, results_hi[5][1]);
        result = _mm_xor_si128(result, results_hi[6][1]);
        result = _mm_xor_si128(result, results_hi[7][1]);
        _mm_storeu_si128((__m128i *)out+1, result);
        return;
}

#endif // __AVX2__

} // extern "C"

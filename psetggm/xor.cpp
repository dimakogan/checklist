#include <cstdint>
#include <cstdio>
#include <cstring>
#include <immintrin.h>
#include <smmintrin.h>

extern "C"
{

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

} // extern "C"
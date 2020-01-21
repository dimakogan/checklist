#ifndef PRG_H__
#define PRG_H__

#include "aes.h"

inline void PRG(const AES_KEY & key, const __m128i seed, __m128i * outbuf, const size_t len)
{
  for (size_t i = 0; i < len; ++i)
  {
    outbuf[i] = _mm_xor_si128(seed, _mm_set_epi64x(0, i));
  }
  AES_ecb_encrypt_blks(outbuf, len, &key);
  for (size_t i = 0; i < len; ++i)
  {
    outbuf[i] = _mm_xor_si128(outbuf[i], _mm_set_epi64x(0, i));
    outbuf[i] = _mm_xor_si128(outbuf[i], seed);
  }
}

#endif
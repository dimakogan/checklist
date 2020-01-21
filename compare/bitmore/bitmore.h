#ifndef BITMORE_H__
#define BITMORE_H__

#include <bitset>
#include <iostream>

#include "dpf.h"

namespace bitmore
{

using namespace dpf;

static const __m256i lo4 = _mm256_set1_epi8(0x0f);

template <size_t nbytes_per_word>
struct word
{ 
  static constexpr size_t len256 = std::ceil(nbytes_per_word / static_cast<double>(sizeof(__m256i)));
  __m256i m256[len256];
  inline word<nbytes_per_word> & operator^=(const word<nbytes_per_word> & other)
  {
    for (size_t i = 0; i < len256; ++i) m256[i] = _mm256_xor_si256(m256[i], other.m256[i]);
    return *this;
  }
};

template <size_t nbytes_per_word, size_t nwords_per_row>
struct record
{
private:
  word<nbytes_per_word> words[nwords_per_row];
public:
  record(word<nbytes_per_word> & word) { std::fill_n(words, nwords_per_row, word); }
  inline constexpr word<nbytes_per_word> & operator[](const size_t j) { return words[j]; }
};

template <size_t nitems, size_t nkeys, uint8_t nservers = static_cast<uint8_t>(std::exp2(nkeys))>
struct query
{
  typedef std::array<dpf_key<nitems>, nkeys> request;

  dpf_key<nitems> dpfkey[nkeys][2];
  std::bitset<nkeys> queries[nservers];

  inline query(AES_KEY & aeskey, size_t point)
  {
  	std::bitset<nkeys> permutation;
    for (size_t i = 0; i < nkeys; ++i) permutation.set(i, dpf::gen(aeskey, point, dpfkey[i]));
    //std::cout << "\npermutation: " << permutation << std::endl;

    if constexpr(std::exp2(nkeys) != static_cast<double>(nservers))
    {
      constexpr size_t remainder = nkeys % 64;

      // bitset is a bit array that keeps track of which permutations we've generated so far
      std::bitset<nservers> mask;
      while (!mask.all())
      {
        std::bitset<nkeys> rnd;
        double modulo = 0.0;
        double shift_64 = std::fmod(std::exp2(64), static_cast<double>(nservers));

        uint64_t rbuf;
        for (size_t i = 0; i < nkeys / 64; ++i)
        {
          arc4random_buf(&rbuf, sizeof(uint64_t));
          (rnd <<= 64) |= rbuf;
          modulo = std::fmod(modulo*shift_64+static_cast<double>(rbuf % nservers), static_cast<double>(nservers));
        }
        if constexpr(remainder)
        {
          arc4random_buf(&rbuf, sizeof(uint64_t));
          rbuf &= (1ULL << remainder) - 1;
          (rnd <<= remainder) |= rbuf;
          double rem_d = std::fmod(static_cast<double>(1ULL << remainder), static_cast<double>(nservers));
          modulo = std::fmod(modulo*rem_d+static_cast<double>(rbuf % nservers), static_cast<double>(nservers));
        }
        uint64_t modulo_ull = static_cast<uint64_t>(modulo);

        if (!mask.test(modulo_ull)) { queries[modulo_ull] = permutation ^ rnd; mask.set(modulo_ull); }
      }
    }
    else
    {
      for (size_t i = 0; i < nservers; ++i) queries[i] = permutation ^ std::bitset<nkeys>(i);
    }
  }

  inline request get_request(uint8_t word)
  {
    request r;
    for (size_t i = 0; i < nkeys; ++i) r[i] = dpfkey[i][queries[word].test(i)];
    return r;
  }
};

template <size_t nkeys, size_t nitems>
inline std::bitset<nkeys> gen(AES_KEY & aeskey, size_t point, dpf_key<nitems> dpfkey[nkeys][2])
{
  std::bitset<nkeys> permutation;
  for (size_t i = 0; i < nkeys; ++i) permutation.set(i, dpf::gen(aeskey, point, dpfkey[i]));
  return permutation;
}


template <uint8_t nservers>
inline __m256i partial_reduce(const __m256i & x)
{
  static const __m256i shuf0 = _mm256_set_epi8(
    (0xf0/nservers)*nservers, (0xe0/nservers)*nservers,
    (0xd0/nservers)*nservers, (0xc0/nservers)*nservers,
    (0xb0/nservers)*nservers, (0xa0/nservers)*nservers,
    (0x90/nservers)*nservers, (0x80/nservers)*nservers,
    (0x70/nservers)*nservers, (0x60/nservers)*nservers,
    (0x50/nservers)*nservers, (0x40/nservers)*nservers,
    (0x30/nservers)*nservers, (0x20/nservers)*nservers,
    (0x10/nservers)*nservers, (0x00/nservers)*nservers,
    (0xf0/nservers)*nservers, (0xe0/nservers)*nservers,
    (0xd0/nservers)*nservers, (0xc0/nservers)*nservers,
    (0xb0/nservers)*nservers, (0xa0/nservers)*nservers,
    (0x90/nservers)*nservers, (0x80/nservers)*nservers,
    (0x70/nservers)*nservers, (0x60/nservers)*nservers,
    (0x50/nservers)*nservers, (0x40/nservers)*nservers,
    (0x30/nservers)*nservers, (0x20/nservers)*nservers,
    (0x10/nservers)*nservers, (0x00/nservers)*nservers
  );

  return _mm256_sub_epi8(x, _mm256_shuffle_epi8(shuf0, _mm256_and_si256(_mm256_srli_epi16(x, 4), lo4)));
}

template <uint8_t nservers>
inline __m256i final_reduce(__m256i & x)
{
  static constexpr uint8_t lg = static_cast<uint8_t>(std::ceil(std::log2(nservers)+1));
  if constexpr(lg >= 5)
  {
    uint8_t * y = reinterpret_cast<uint8_t *>(&x);
    for (size_t i = 0; i < 32; ++i) y[i] %= nservers;
    return x;
  }
  else if constexpr(lg == 5)
  {
    static const __m256i y = _mm256_set1_epi8(static_cast<uint8_t>((0x10 / nservers) * nservers));
    return _mm256_sub_epi8(x, _mm256_and_si256(_mm256_cmpgt_epi8(x, lo4), y));
  }
  else
  {
    static const __m256i shuf = _mm256_set_epi8(
      (0x0f/nservers)*nservers, (0x0e/nservers)*nservers,
      (0x0d/nservers)*nservers, (0x0c/nservers)*nservers,
      (0x0b/nservers)*nservers, (0x0a/nservers)*nservers,
      (0x09/nservers)*nservers, (0x08/nservers)*nservers,
      (0x07/nservers)*nservers, (0x06/nservers)*nservers,
      (0x05/nservers)*nservers, (0x04/nservers)*nservers,
      (0x03/nservers)*nservers, (0x02/nservers)*nservers,
      (0x01/nservers)*nservers, (0x00/nservers)*nservers,
      (0x0f/nservers)*nservers, (0x0e/nservers)*nservers,
      (0x0d/nservers)*nservers, (0x0c/nservers)*nservers,
      (0x0b/nservers)*nservers, (0x0a/nservers)*nservers,
      (0x09/nservers)*nservers, (0x08/nservers)*nservers,
      (0x07/nservers)*nservers, (0x06/nservers)*nservers,
      (0x05/nservers)*nservers, (0x04/nservers)*nservers,
      (0x03/nservers)*nservers, (0x02/nservers)*nservers,
      (0x01/nservers)*nservers, (0x00/nservers)*nservers
    );
  
    return _mm256_sub_epi8(x, _mm256_shuffle_epi8(shuf, _mm256_and_si256(x, lo4)));
  }
}

template <size_t nkeys, uint8_t nservers, bool do_reduce = static_cast<bool>(nservers < (std::exp2(nkeys)))>
inline static void splice(const __m128i mask[nkeys], __m1024i out)
{
  const static __m256i shuffle = _mm256_setr_epi64x(0x0000000000000000,
    0x0101010101010101, 0x0202020202020202, 0x0303030303030303);
  const static __m256i bit_mask = _mm256_set1_epi64x(0x7fbfdfeff7fbfdfe);
  const static __m256i ff = _mm256_set1_epi8(0xff);
  const __m256i shift[] = { _mm256_set1_epi8(0x01), _mm256_set1_epi8(0x02),
    _mm256_set1_epi8(0x04), _mm256_set1_epi8(0x08), _mm256_set1_epi8(0x10),
    _mm256_set1_epi8(0x20), _mm256_set1_epi8(0x40), _mm256_set1_epi8(0x80) };

  constexpr uint8_t lgbit = 1U << static_cast<uint8_t>(8 - std::ceil(std::log2(nservers)));

  for (size_t i = 0; i < nkeys; ++i)
  {
    const __m256i rem = (do_reduce) ? _mm256_set1_epi8(static_cast<uint8_t>(std::fmod(std::exp2(i),nservers))) : shift[i];
    const uint32_t * maski = (uint32_t *)&mask[i];
    for (int j = 0; j < 4; j++)
    {
      out[j] = _mm256_add_epi8(out[j], // accumulate into result
        _mm256_and_si256(rem,
          _mm256_cmpeq_epi8( // 0xff if byte is 0xff; else 0x00
            _mm256_or_si256( // set bits not possibly set by shuffle
              _mm256_shuffle_epi8( // shuffle 32 bits into 32 bytes
                _mm256_set1_epi32(maski[j]),
              shuffle),
            bit_mask),
          ff)
        )
      );
      if constexpr(do_reduce)
      {
        if (i & lgbit) out[j] = partial_reduce<nservers>(out[j]);
      }
    }
  }
  if constexpr(do_reduce)
  {
    for (uint8_t j = 0; j < 4; ++j)
    {
      if constexpr(nkeys % lgbit) out[j] = partial_reduce<nservers>(out[j]);
      out[j] = final_reduce<nservers>(out[j]);
    }
  }
}

template <size_t nkeys, size_t nitems, uint8_t nservers>
inline void evalfull(AES_KEY & aeskey, typename query<nitems, nkeys, nservers>::request & dpfkey,
  __m128i ** s, uint8_t ** t, uint8_t * output)
{
  constexpr size_t depth = dpf_key<nitems>::depth;
  constexpr size_t output_length = dpf_key<nitems>::output_length;

  __m128i child[2];
  uint8_t ts[2];
  for (size_t l = 0; l < nkeys; ++l)
  {
    int curlayer = depth % 2;

    __m128i * s_[2] = { s[l], s[l] + output_length/2 };
    uint8_t * t_[2] = { t[l], t[l] + output_length/2 };

    s_[curlayer][0] = dpfkey[l].root;
    t_[curlayer][0] = _mm_getlsb_si128(dpfkey[l].root);

    for (size_t i = 0; i < depth; ++i)
    {
      curlayer = 1 - curlayer;
      const size_t itemnumber = std::max(output_length >> (depth-i), 1UL);
      for (size_t j = 0; j < itemnumber; ++j)
      {
        expand(aeskey, s_[1-curlayer][j], child, ts);
        s_[curlayer][2*j] = _mm_xorif_si128(child[L], dpfkey[l].cw[i], t_[1-curlayer][j]);
        t_[curlayer][2*j] = ts[L] ^ dpfkey[l].t[i][L] & t_[1-curlayer][j];
        if (2*j+1 < 2*itemnumber)
        {
          s_[curlayer][2*j+1] = _mm_xorif_si128(child[R], dpfkey[l].cw[i], t_[1-curlayer][j]);
          t_[curlayer][2*j+1] = ts[R] ^ dpfkey[l].t[i][R] & t_[1-curlayer][j];
        }
      }
    }
  }

  memset(output, 0, nitems * sizeof(uint8_t));
  __m128i tmp[nkeys];
  __m1024i * output1024 = reinterpret_cast<__m1024i *>(output);
  for (size_t j = 0; j < output_length; ++j)
  {
    for (size_t l = 0; l < nkeys; ++l)
    {
      tmp[l] = _mm_xorif_si128(s[l][j], dpfkey[l].final, t[l][j]);
    }
    splice<nkeys, nservers>(tmp, output1024[j]);
  }
}

template <size_t nkeys, uint8_t nservers, size_t nitems>
inline uint8_t eval(AES_KEY & aeskey, dpf_key<nitems> dpfkey[nkeys],
  const size_t input)
{
  size_t result = 0;
  for (size_t i = 0; i < nkeys; ++i)
  {
    if (dpf::eval(aeskey, dpfkey[i], input)) result += static_cast<size_t>(std::fmod(std::exp2(i),nservers));
  }
  return static_cast<uint8_t>(result % nservers);
}

}

#endif
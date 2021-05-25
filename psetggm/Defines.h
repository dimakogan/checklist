#pragma once
#include <cstdint>


#include "intrinsics.h"

#include "gsl-lite.hpp"

#define STRINGIZE_DETAIL(x) #x
#define STRINGIZE(x) STRINGIZE_DETAIL(x)
#define LOCATION __FILE__ ":" STRINGIZE(__LINE__)

template<typename T> using span = gsl::span<T>;

typedef  __m128i block;
typedef  union {
    __m128i reg;
    uint8_t arr[16];
} reg_arr_union;

inline block toBlock(const uint8_t* in) { return _mm_set_epi64x(((uint64_t*)in)[1], ((uint64_t*)in)[0]);}
//inline void fromBlock(uint8_t* out, const block& in) {vst1q_u8(out, in);}
inline block dupUint64(uint64_t val) { return _mm_set_epi64x(val, val);}

inline bool eq(const block& lhs, const block& rhs)
{
  block neq = _mm_xor_si128(lhs, rhs);
  return _mm_test_all_zeros(neq, neq) != 0;
}

inline bool neq(const block& lhs, const block& rhs)
{
  block neq = _mm_xor_si128(lhs, rhs);
  return _mm_test_all_zeros(neq, neq) == 0;
}

inline bool is_zero(const block& b) {
    return _mm_test_all_zeros(b, b);
}

extern const block ZeroBlock;
extern const block LSBBlock;
extern const block MSBBlock;
extern const block AllOneBlock;
extern const block TestBlock;


void split(const std::string &s, char delim, std::vector<std::string> &elems);
std::vector<std::string> split(const std::string &s, char delim);

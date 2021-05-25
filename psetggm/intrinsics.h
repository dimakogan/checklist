#ifdef  __amd64__

#include <emmintrin.h>
#include <immintrin.h>
#include <smmintrin.h>
#include <wmmintrin.h>
#include <xmmintrin.h>

#else 

#include "sse2neon.h"

#endif // __amd64__
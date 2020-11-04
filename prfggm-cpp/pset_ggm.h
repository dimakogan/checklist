#include <cstdint>
#include <immintrin.h>
#include <smmintrin.h>
#include <vector>

class SetGenerator {
    public:
        SetGenerator(uint32_t univ_size, uint32_t set_size);

        void Eval(__m128i seed, uint32_t *out);

    private:
        void tree_eval_all(uint32_t key_pos, uint32_t height, uint32_t* out);
        void expand(uint32_t key_pos);
    private:
        uint32_t _univ_size, _set_size, _height;
        std::vector<__m128i> _path_key;
};
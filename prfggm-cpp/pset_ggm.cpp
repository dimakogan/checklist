#include "AES.h"
#include "pset_ggm.h"

int height(int v) {
    unsigned int r = 0; // r will be lg(v)
    --v;

    while (v >>= 1) 
    {
        r++;
    }
    return r+1;
}

SetGenerator::SetGenerator(uint32_t univ_size, uint32_t set_size) 
    :  
        _univ_size(univ_size),  
        _set_size(set_size),
        _height(height(set_size)),
        _path_key(2*(_height+1)) {}

void SetGenerator::Eval(__m128i seed, uint32_t *out) {
    _path_key[0] = seed;
    tree_eval_all(0, _height, out);
}

void SetGenerator::tree_eval_all(uint32_t key_pos, uint32_t height, uint32_t* out) {
    uint32_t node = 0;
    while (true) {
        if (height == 0) {
            out[node] = *(uint32_t*)(&_path_key[key_pos]) % _univ_size;
            bool is_right = true;
            // while 'is right child', go up
            while (node&1 == 1) {
                ++height;
                key_pos -= 1;
                node >>= 1;
            }
            if (height >= _height) {
                return;
            }
            // move to right sibling
            node += 1;
            key_pos -= 1;

            if ((node << height) >=  _set_size) {
                return;
            }

            continue;
        }
        expand(key_pos);
        node <<= 1;
        --height;
        // first go to left child
        key_pos += 2;
    }
}

__m128i one = _mm_setr_epi32(0, 0, 0, 1);

void SetGenerator::expand(uint32_t key_pos) {
    mAesFixedKey.encryptECB(_path_key[key_pos], _path_key[key_pos+1]);
    _path_key[key_pos+1] = _mm_xor_si128(_path_key[key_pos],_path_key[key_pos+1]);

    _path_key[key_pos+2] = _mm_xor_si128(_path_key[key_pos], one);
    mAesFixedKey.encryptECB(_path_key[key_pos], _path_key[key_pos+2]);
    _path_key[key_pos+2] = _mm_xor_si128(_path_key[key_pos],_path_key[key_pos+2]);
    _path_key[key_pos+2] = _mm_xor_si128(_path_key[key_pos+2], one);
}
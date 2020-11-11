#include "AES.h"
#include "pset_ggm.h"
#include <vector>

extern "C" {

int get_height(unsigned int v) {
    unsigned int r = 0; // r will be lg(v)
    --v;

    while (v >>= 1) 
    {
        r++;
    }
    return r+1;
}

typedef struct generator {
    unsigned int univ_size, set_size;
    __m128i *keys, *tmp;
} generator;

unsigned int workspace_size(unsigned int univ_size, unsigned int set_size) {
    uint32_t height = get_height(set_size); 
    return sizeof(generator) + 2*(1<<height)*sizeof(__m128i)+32;
}

generator* pset_ggm_init(unsigned int univ_size, unsigned int set_size, uint8_t* workspace) {
    auto gen = (generator*)workspace;
    gen->univ_size = univ_size;
    gen->set_size = set_size;
    gen->keys = (__m128i*)(workspace + sizeof(generator));
    // align pointer
    gen->keys = (__m128i*)((((size_t)gen->keys-1)/16+1)*16);

    uint32_t height = get_height(set_size); 
    gen->tmp = gen->keys + (1<<height);

    return gen;
}

const __m128i one = _mm_setr_epi32(0, 0, 0, 1);

void expand(const __m128i& in, __m128i* out) {
    out[1] = _mm_xor_si128(in, one);
    mAesFixedKey.encryptECB(in, out[0]);
    mAesFixedKey.encryptECB(in, out[1]);
    out[0] = _mm_xor_si128(in, out[0]);
    out[1] = _mm_xor_si128(in, out[1]);
    out[1] = _mm_xor_si128(out[1], one);
}


void tree_eval_all2(generator* gen, __m128i seed, long long unsigned int* elems) {
    uint32_t max_depth = get_height(gen->set_size); 
    int num_layers = max_depth - 2;

    
    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;


    _mm_store_si128(keys, seed);

    for (int depth = 0; depth < num_layers; depth++) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
    }

    const uint32_t* keys_as_elems = (uint32_t*)gen->keys;
    for (int i = 0; i < gen->set_size; i++) {
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        elems[i] =  ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32; //gen->keys[i] % gen->univ_size; //
    }
}

void pset_ggm_eval(generator* gen, const uint8_t* seed, long long unsigned* elems) {
    tree_eval_all2(gen, toBlock(seed), elems);
}

unsigned int pset_buffer_size(const generator* gen) {
    uint32_t height = get_height(gen->set_size); 
    if (height < 2) {
        return sizeof(__m128i);
    }
    return sizeof(__m128i)*(height-1);
}

void pset_ggm_punc(generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset) {
    __m128i* pset_keys = (__m128i*)pset;

    __m128i* keys = (__m128i*)gen->keys;
    __m128i* tmp = (__m128i*)gen->tmp;
    __m128i key = toBlock(seed);

    int depth = 0;
    uint32_t height = get_height(gen->set_size); 

    while (height > 2) {
        _mm_store_si128(tmp, key);
        key = _mm_xor_si128(key, one);
        _mm_store_si128(tmp + 1, key);
        mAesFixedKey.encryptECBBlocks(tmp, 2, keys);
        keys[1] = _mm_xor_si128(keys[1], key);
        key = _mm_xor_si128(key, one);
        keys[0] = _mm_xor_si128(keys[0], key);

        if ((pos & (1<<(height-1))) != 0) {
            pset_keys[depth] = keys[0];
            key = keys[1];
        } else {
            pset_keys[depth] = keys[1];
            key = keys[0];
        }
        depth++;
        height--;
    }
    pset_keys[depth] = key;
    uint32_t* last_key = (uint32_t*)&pset_keys[depth];
    switch (pos & 0b11) {
        case 0:
            last_key[0] = 0;
            break;
        case 1:
            last_key[1] = 0;
            break;
        case 2:
            last_key[2] = 0;
            break;
        case 3:
            last_key[3] = 0;
            break;
    }
}

void pset_ggm_eval_punc(generator* gen, const uint8_t* pset, unsigned int pos, long long unsigned int* elems) {
    uint32_t height = get_height(gen->set_size); 

    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;

    const __m128i* pset_keys = (const __m128i*)pset;
    
    int depth = 0;
    while (height > 2) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
        height--;
        keys[(pos >> height)^1] = pset_keys[depth];
        depth++;
    }
    
    keys[(pos >> height)] = pset_keys[depth];
    
    size_t out_pos = 0;
    uint32_t* keys_as_elems = (uint32_t*)keys;
    for (int i = 0; i < gen->set_size; i++) {
        if (i == pos) {
            continue;
        }
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        elems[out_pos] = ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32; 
        out_pos++;
    }  
}

}
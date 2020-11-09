#include "AES.h"
#include "pset_ggm.h"
#include <vector>

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
    std::vector<uint32_t> keys, tmp;
} generator;


generator* init_generator(unsigned int univ_size, unsigned int set_size) {
    generator* gen = new(generator);
    gen->univ_size = univ_size;
    gen->set_size = set_size;

    uint32_t height = get_height(set_size); 
    gen->keys.resize(1<<height);
    gen->tmp.resize(1<<height);

    return gen;
}

void free_generator(generator* gen) {
    delete(gen);
}

void free_pset(uint8_t* pset) {
    free(pset);
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



void tree_eval_all(unsigned int	univ_size, unsigned int set_size, __m128i seed, uint32_t* out) {
    uint32_t key_pos = 0;
    uint32_t max_height = get_height(set_size); 
    uint32_t height = max_height;
    std::vector<__m128i> path_key(2*(max_height+1));
    path_key[0] = seed;
    uint32_t node = 0;
    while (true) {
        if (height == 0) {
            out[node] = *(uint32_t*)(&path_key[key_pos]) % univ_size;
            bool is_right = true;
            // while 'is right child', go up
            while (node&1 == 1) {
                ++height;
                key_pos -= 1;
                node >>= 1;
            }
            if (height >= max_height) {
                return;
            }
            // move to right sibling
            node += 1;
            key_pos -= 1;

            if ((node << height) >=  set_size) {
                return;
            }

            continue;
        }
        expand(path_key[key_pos], &path_key[key_pos+1]);
        node <<= 1;
        --height;
        // first go to left child
        key_pos += 2;
    }
}

void tree_eval_all2(generator* gen, __m128i seed, uint32_t* out) {
    uint32_t max_depth = get_height(gen->set_size); 
    uint32_t num_layers = max_depth - 2;

    
    __m128i* keys = (__m128i*)gen->keys.data();
    // Out needs to be twice as large as output to be used as workspace.
    __m128i* tmp =  (__m128i*)gen->tmp.data();


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

    for (int i = 0; i < gen->set_size; i++) {
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        out[i] = gen->keys[i] % gen->univ_size; // ((uint64_t) gen->keys[i] * (uint64_t) gen->univ_size) >> 32; //out[i] % univ_size;
    }
}


extern "C" {

void pset_ggm_eval(generator* gen, const uint8_t* seed, uint32_t* out) {
    tree_eval_all2(gen, toBlock(seed), out);
}

uint8_t* pset_ggm_punc(generator* gen, const uint8_t* seed, unsigned int pos, unsigned int* out_size) {
    uint32_t height = get_height(gen->set_size); 

    *out_size = 16*(height-1);
    __m128i* pset = (__m128i*)malloc(16*(height-1));

    __m128i* keys = (__m128i*)gen->keys.data();
    __m128i* tmp = (__m128i*)gen->tmp.data();
    __m128i key = toBlock(seed);

    int depth = 0;

    while (height > 2) {
        _mm_store_si128(tmp, key);
        key = _mm_xor_si128(key, one);
        _mm_store_si128(tmp + 1, key);
        mAesFixedKey.encryptECBBlocks(tmp, 2, keys);
        keys[1] = _mm_xor_si128(keys[1], key);
        key = _mm_xor_si128(key, one);
        keys[0] = _mm_xor_si128(keys[0], key);

        if ((pos & (1<<(height-1))) != 0) {
            pset[depth] = keys[0];
            key = keys[1];
        } else {
            pset[depth] = keys[1];
            key = keys[0];
        }
        depth++;
        height--;
    }
    pset[depth] = key;
    uint32_t* last_key = (uint32_t*)&pset[depth];
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
    
    return (uint8_t*)pset;
}

void pset_ggm_eval_punc(generator* gen, const uint8_t* pset, unsigned int pos, uint32_t* out) {
    uint32_t height = get_height(gen->set_size); 

    __m128i* keys = (__m128i*)gen->keys.data();
    // Out needs to be twice as large as output to be used as workspace.
    __m128i* tmp =  (__m128i*)gen->tmp.data();

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
    for (int i = 0; i < gen->set_size; i++) {
        // if (i == pos) {
        //     continue;
        // }
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        out[out_pos] = gen->keys[i] % gen->univ_size; //((uint64_t) gen->keys[i] * (uint64_t) gen->univ_size) >> 32; 
        out_pos++;
    }  
}

}
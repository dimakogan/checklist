#pragma once

#include <cstdint>

#include "Defines.h"


#include <wmmintrin.h>
#include <xmmintrin.h>

class AES {
public:
    AES();
    AES(const block& key);
    AES(const uint8_t* key);

    void setKey(const block& key);
    void setKey(const uint8_t* key);

    void encryptECB(const block& plaintext, block& ciphertext) const;
    block encryptECB(const block& plaintext) const {
        block tmp;
        encryptECB(plaintext, tmp);
        return tmp;
    }
    void encryptECB_MMO(const block& plaintext, block& ciphertext) const;
    block encryptECB_MMO(const block& plaintext) const {
        block tmp;
        encryptECB_MMO(plaintext, tmp);
        return tmp;
    }

    void decryptECB(const block& ciphertext, block& plaintext) const;
    block decryptECB(const block& ciphertext) const {
        block tmp;
        decryptECB(ciphertext, tmp);
        return tmp;
    }
    void encryptECBBlocks(const block* plaintexts, uint64_t blockLength, block* ciphertexts) const;
    void encryptECB_MMO_Blocks(const block* plaintexts, uint64_t blockLength, block* ciphertexts) const;

    void encryptCTR(uint64_t baseIdx, uint64_t blockLength, block * ciphertext) const;
    block key;
private:
    block mRoundKeysEnc[11];
    block mRoundKeysDec[11];
};

// An AES instance with a fixed and public key
extern const AES mAesFixedKey;
extern const AES mAesFixedKey2;


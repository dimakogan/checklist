#include "AES.h"
#include <cassert>


const uint8_t fixed_key[16] = {36,156,50,234,92,230,49,9,174,170,205,160,98,236,29,243};
const AES mAesFixedKey(fixed_key);
const uint8_t fixed_key2[16] = {209, 12, 199, 173, 29, 74, 44, 128, 194, 224, 14, 44, 2, 201, 110, 28};
const AES mAesFixedKey2(fixed_key2);


block keyGenHelper(block key, block keyRcon)
{
    keyRcon = _mm_shuffle_epi32(keyRcon, _MM_SHUFFLE(3, 3, 3, 3));
    key = _mm_xor_si128(key, _mm_slli_si128(key, 4));
    key = _mm_xor_si128(key, _mm_slli_si128(key, 4));
    key = _mm_xor_si128(key, _mm_slli_si128(key, 4));
    return _mm_xor_si128(key, keyRcon);
}

AES::AES() {
    uint8_t zerokey[] = {0,0,0,0, 0,0,0,0, 0,0,0,0 , 0,0,0,0};
    setKey(toBlock(zerokey));
}

AES::AES(const block& key) {
    setKey(key);
}
AES::AES(const uint8_t* key) {
    setKey(key);
}

void AES::setKey(const block& key) {
    mRoundKeysEnc[0] = key;
    mRoundKeysEnc[1] = keyGenHelper(mRoundKeysEnc[0], _mm_aeskeygenassist_si128(mRoundKeysEnc[0], 0x01));
    mRoundKeysEnc[2] = keyGenHelper(mRoundKeysEnc[1], _mm_aeskeygenassist_si128(mRoundKeysEnc[1], 0x02));
    mRoundKeysEnc[3] = keyGenHelper(mRoundKeysEnc[2], _mm_aeskeygenassist_si128(mRoundKeysEnc[2], 0x04));
    mRoundKeysEnc[4] = keyGenHelper(mRoundKeysEnc[3], _mm_aeskeygenassist_si128(mRoundKeysEnc[3], 0x08));
    mRoundKeysEnc[5] = keyGenHelper(mRoundKeysEnc[4], _mm_aeskeygenassist_si128(mRoundKeysEnc[4], 0x10));
    mRoundKeysEnc[6] = keyGenHelper(mRoundKeysEnc[5], _mm_aeskeygenassist_si128(mRoundKeysEnc[5], 0x20));
    mRoundKeysEnc[7] = keyGenHelper(mRoundKeysEnc[6], _mm_aeskeygenassist_si128(mRoundKeysEnc[6], 0x40));
    mRoundKeysEnc[8] = keyGenHelper(mRoundKeysEnc[7], _mm_aeskeygenassist_si128(mRoundKeysEnc[7], 0x80));
    mRoundKeysEnc[9] = keyGenHelper(mRoundKeysEnc[8], _mm_aeskeygenassist_si128(mRoundKeysEnc[8], 0x1B));
    mRoundKeysEnc[10] = keyGenHelper(mRoundKeysEnc[9], _mm_aeskeygenassist_si128(mRoundKeysEnc[9], 0x36));
}

void AES::setKey(const uint8_t* key) {
   setKey(toBlock(key));
}

void AES::encryptECB(const block& plaintext, block& ciphertext) const {
    ciphertext = _mm_xor_si128(plaintext, mRoundKeysEnc[0]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[1]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[2]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[3]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[4]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[5]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[6]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[7]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[8]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[9]);
    ciphertext = _mm_aesenclast_si128(ciphertext, mRoundKeysEnc[10]);
}

void AES::encryptECB_MMO(const block& plaintext, block& ciphertext) const {
    ciphertext = _mm_xor_si128(plaintext, mRoundKeysEnc[0]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[1]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[2]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[3]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[4]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[5]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[6]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[7]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[8]);
    ciphertext = _mm_aesenc_si128(ciphertext, mRoundKeysEnc[9]);
    ciphertext = _mm_aesenclast_si128(ciphertext, mRoundKeysEnc[10]);
    ciphertext = _mm_xor_si128(ciphertext, plaintext);
}

void AES::encryptECBBlocks(const block* plaintexts, uint64_t blockLength, block* ciphertexts) const {

    const uint64_t step = 8;
    uint64_t idx = 0;
    uint64_t length = blockLength - blockLength % step;

    //std::array<block, step> temp;
    block temp[step];

    for (; idx < length; idx += step)
    {
        temp[0] = _mm_xor_si128(plaintexts[idx + 0], mRoundKeysEnc[0]);
        temp[1] = _mm_xor_si128(plaintexts[idx + 1], mRoundKeysEnc[0]);
        temp[2] = _mm_xor_si128(plaintexts[idx + 2], mRoundKeysEnc[0]);
        temp[3] = _mm_xor_si128(plaintexts[idx + 3], mRoundKeysEnc[0]);
        temp[4] = _mm_xor_si128(plaintexts[idx + 4], mRoundKeysEnc[0]);
        temp[5] = _mm_xor_si128(plaintexts[idx + 5], mRoundKeysEnc[0]);
        temp[6] = _mm_xor_si128(plaintexts[idx + 6], mRoundKeysEnc[0]);
        temp[7] = _mm_xor_si128(plaintexts[idx + 7], mRoundKeysEnc[0]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[1]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[1]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[1]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[1]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[1]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[1]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[1]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[1]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[2]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[2]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[2]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[2]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[2]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[2]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[2]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[2]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[3]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[3]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[3]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[3]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[3]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[3]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[3]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[3]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[4]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[4]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[4]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[4]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[4]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[4]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[4]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[4]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[5]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[5]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[5]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[5]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[5]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[5]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[5]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[5]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[6]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[6]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[6]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[6]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[6]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[6]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[6]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[6]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[7]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[7]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[7]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[7]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[7]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[7]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[7]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[7]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[8]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[8]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[8]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[8]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[8]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[8]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[8]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[8]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[9]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[9]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[9]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[9]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[9]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[9]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[9]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[9]);

        ciphertexts[idx + 0] = _mm_aesenclast_si128(temp[0], mRoundKeysEnc[10]);
        ciphertexts[idx + 1] = _mm_aesenclast_si128(temp[1], mRoundKeysEnc[10]);
        ciphertexts[idx + 2] = _mm_aesenclast_si128(temp[2], mRoundKeysEnc[10]);
        ciphertexts[idx + 3] = _mm_aesenclast_si128(temp[3], mRoundKeysEnc[10]);
        ciphertexts[idx + 4] = _mm_aesenclast_si128(temp[4], mRoundKeysEnc[10]);
        ciphertexts[idx + 5] = _mm_aesenclast_si128(temp[5], mRoundKeysEnc[10]);
        ciphertexts[idx + 6] = _mm_aesenclast_si128(temp[6], mRoundKeysEnc[10]);
        ciphertexts[idx + 7] = _mm_aesenclast_si128(temp[7], mRoundKeysEnc[10]);
    }

    for (; idx < blockLength; ++idx)
    {
        ciphertexts[idx] = _mm_xor_si128(plaintexts[idx], mRoundKeysEnc[0]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[1]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[2]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[3]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[4]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[5]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[6]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[7]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[8]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[9]);
        ciphertexts[idx] = _mm_aesenclast_si128(ciphertexts[idx], mRoundKeysEnc[10]);
    }
}

void AES::encryptECB_MMO_Blocks(const block* plaintexts, uint64_t blockLength, block* ciphertexts) const {

    const uint64_t step = 8;
    uint64_t idx = 0;
    uint64_t length = blockLength - blockLength % step;

    //std::array<block, step> temp;
    block temp[step];

    for (; idx < length; idx += step)
    {
        temp[0] = _mm_xor_si128(plaintexts[idx + 0], mRoundKeysEnc[0]);
        temp[1] = _mm_xor_si128(plaintexts[idx + 1], mRoundKeysEnc[0]);
        temp[2] = _mm_xor_si128(plaintexts[idx + 2], mRoundKeysEnc[0]);
        temp[3] = _mm_xor_si128(plaintexts[idx + 3], mRoundKeysEnc[0]);
        temp[4] = _mm_xor_si128(plaintexts[idx + 4], mRoundKeysEnc[0]);
        temp[5] = _mm_xor_si128(plaintexts[idx + 5], mRoundKeysEnc[0]);
        temp[6] = _mm_xor_si128(plaintexts[idx + 6], mRoundKeysEnc[0]);
        temp[7] = _mm_xor_si128(plaintexts[idx + 7], mRoundKeysEnc[0]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[1]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[1]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[1]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[1]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[1]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[1]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[1]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[1]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[2]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[2]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[2]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[2]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[2]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[2]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[2]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[2]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[3]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[3]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[3]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[3]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[3]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[3]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[3]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[3]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[4]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[4]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[4]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[4]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[4]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[4]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[4]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[4]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[5]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[5]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[5]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[5]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[5]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[5]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[5]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[5]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[6]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[6]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[6]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[6]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[6]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[6]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[6]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[6]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[7]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[7]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[7]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[7]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[7]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[7]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[7]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[7]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[8]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[8]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[8]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[8]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[8]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[8]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[8]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[8]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[9]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[9]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[9]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[9]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[9]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[9]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[9]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[9]);

        temp[0] = _mm_aesenclast_si128(temp[0], mRoundKeysEnc[10]);
        temp[1] = _mm_aesenclast_si128(temp[1], mRoundKeysEnc[10]);
        temp[2] = _mm_aesenclast_si128(temp[2], mRoundKeysEnc[10]);
        temp[3] = _mm_aesenclast_si128(temp[3], mRoundKeysEnc[10]);
        temp[4] = _mm_aesenclast_si128(temp[4], mRoundKeysEnc[10]);
        temp[5] = _mm_aesenclast_si128(temp[5], mRoundKeysEnc[10]);
        temp[6] = _mm_aesenclast_si128(temp[6], mRoundKeysEnc[10]);
        temp[7] = _mm_aesenclast_si128(temp[7], mRoundKeysEnc[10]);

    	ciphertexts[idx + 0] = _mm_xor_si128(temp[0], plaintexts[0]);
    	ciphertexts[idx + 1] = _mm_xor_si128(temp[1], plaintexts[1]);
    	ciphertexts[idx + 2] = _mm_xor_si128(temp[2], plaintexts[2]);
    	ciphertexts[idx + 3] = _mm_xor_si128(temp[3], plaintexts[3]);
    	ciphertexts[idx + 4] = _mm_xor_si128(temp[4], plaintexts[4]);
    	ciphertexts[idx + 5] = _mm_xor_si128(temp[5], plaintexts[5]);
    	ciphertexts[idx + 6] = _mm_xor_si128(temp[6], plaintexts[6]);
    	ciphertexts[idx + 7] = _mm_xor_si128(temp[7], plaintexts[7]);
    }

    for (; idx < blockLength; ++idx)
    {
        ciphertexts[idx] = _mm_xor_si128(plaintexts[idx], mRoundKeysEnc[0]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[1]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[2]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[3]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[4]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[5]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[6]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[7]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[8]);
        ciphertexts[idx] = _mm_aesenc_si128(ciphertexts[idx], mRoundKeysEnc[9]);
        ciphertexts[idx] = _mm_aesenclast_si128(ciphertexts[idx], mRoundKeysEnc[10]);
        ciphertexts[idx] = _mm_xor_si128(plaintexts[idx], ciphertexts[idx]);
    }
}

void AES::encryptCTR(uint64_t baseIdx, uint64_t blockLength, block * ciphertext) const {

    const uint64_t step = 8;
    uint64_t idx = 0;
    uint64_t length = blockLength - blockLength % step;

    //std::array<block, step> temp;
    block temp[step];

    for (; idx < length; idx += step, baseIdx += step)
    {
        temp[0] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 0), mRoundKeysEnc[0]);
        temp[1] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 1), mRoundKeysEnc[0]);
        temp[2] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 2), mRoundKeysEnc[0]);
        temp[3] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 3), mRoundKeysEnc[0]);
        temp[4] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 4), mRoundKeysEnc[0]);
        temp[5] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 5), mRoundKeysEnc[0]);
        temp[6] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 6), mRoundKeysEnc[0]);
        temp[7] = _mm_xor_si128(_mm_set1_epi64x(baseIdx + 7), mRoundKeysEnc[0]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[1]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[1]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[1]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[1]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[1]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[1]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[1]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[1]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[2]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[2]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[2]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[2]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[2]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[2]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[2]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[2]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[3]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[3]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[3]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[3]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[3]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[3]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[3]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[3]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[4]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[4]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[4]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[4]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[4]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[4]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[4]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[4]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[5]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[5]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[5]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[5]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[5]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[5]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[5]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[5]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[6]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[6]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[6]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[6]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[6]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[6]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[6]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[6]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[7]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[7]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[7]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[7]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[7]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[7]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[7]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[7]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[8]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[8]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[8]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[8]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[8]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[8]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[8]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[8]);

        temp[0] = _mm_aesenc_si128(temp[0], mRoundKeysEnc[9]);
        temp[1] = _mm_aesenc_si128(temp[1], mRoundKeysEnc[9]);
        temp[2] = _mm_aesenc_si128(temp[2], mRoundKeysEnc[9]);
        temp[3] = _mm_aesenc_si128(temp[3], mRoundKeysEnc[9]);
        temp[4] = _mm_aesenc_si128(temp[4], mRoundKeysEnc[9]);
        temp[5] = _mm_aesenc_si128(temp[5], mRoundKeysEnc[9]);
        temp[6] = _mm_aesenc_si128(temp[6], mRoundKeysEnc[9]);
        temp[7] = _mm_aesenc_si128(temp[7], mRoundKeysEnc[9]);

        ciphertext[idx + 0] = _mm_aesenclast_si128(temp[0], mRoundKeysEnc[10]);
        ciphertext[idx + 1] = _mm_aesenclast_si128(temp[1], mRoundKeysEnc[10]);
        ciphertext[idx + 2] = _mm_aesenclast_si128(temp[2], mRoundKeysEnc[10]);
        ciphertext[idx + 3] = _mm_aesenclast_si128(temp[3], mRoundKeysEnc[10]);
        ciphertext[idx + 4] = _mm_aesenclast_si128(temp[4], mRoundKeysEnc[10]);
        ciphertext[idx + 5] = _mm_aesenclast_si128(temp[5], mRoundKeysEnc[10]);
        ciphertext[idx + 6] = _mm_aesenclast_si128(temp[6], mRoundKeysEnc[10]);
        ciphertext[idx + 7] = _mm_aesenclast_si128(temp[7], mRoundKeysEnc[10]);
    }

    for (; idx < blockLength; ++idx, ++baseIdx)
    {
        ciphertext[idx] = _mm_xor_si128(_mm_set1_epi64x(baseIdx), mRoundKeysEnc[0]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[1]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[2]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[3]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[4]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[5]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[6]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[7]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[8]);
        ciphertext[idx] = _mm_aesenc_si128(ciphertext[idx], mRoundKeysEnc[9]);
        ciphertext[idx] = _mm_aesenclast_si128(ciphertext[idx], mRoundKeysEnc[10]);
    }
}

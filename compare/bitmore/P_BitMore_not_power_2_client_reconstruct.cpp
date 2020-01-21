#include <bsd/stdlib.h> // for arc4random_buf
#include <cmath>        // for std::log2 and std::ceil
#include <cstring>      // for std::memcpy
#include <x86intrin.h>

#include <iostream>
#include <string>

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


int main(int argc, char * argv[])
{
	constexpr uint8_t nservers = 5;
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = nservers-1;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	constexpr size_t item = 401;
	
	typedef record<nbytes_per_word, nwords_per_row> record;

	//response word from s+1 th server
	std::string filename = "p_bm_np2_response_from_server_" + std::to_string(nservers-1);
	FILE* file4 = fopen(filename.c_str(), "rb");
	word<nbytes_per_word> qqtemplate = { 0 };
    fread (&qqtemplate, sizeof(qqtemplate), 1, file4);
    fclose(file4);

    //response words from remaining 1 to s th server
	record result = { qqtemplate };
	for (size_t k = 0; k < nservers-1; ++k)
	{	std::string filename2 = "p_bm_np2_response_from_server_" + std::to_string(k);
		FILE* file5 = fopen(filename2.c_str(), "rb");
		word<nbytes_per_word> temp_word = { 0 };
	    fread (&temp_word, sizeof(temp_word), 1, file5);
	    fclose(file5);
		result[k] ^= temp_word;
	}
	return 0;
}

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-result"

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
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = 3;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));

	typedef record<nbytes_per_word, nwords_per_row> record;

	word<nbytes_per_word> zerow = { 0 };
	//response from 1st server
	std::string filename = "c_boyle_response_from_server_0";
	FILE* file4 = fopen(filename.c_str(), "rb");
	record result = { zerow };
    fread (&result, sizeof(result), 1, file4);
    fclose(file4);

    //response from 2nd server
    std::string filename2 = "c_boyle_response_from_server_1";
	FILE* file = fopen(filename2.c_str(), "rb");
	record result1 = { zerow };
    fread (&result1, sizeof(result1), 1, file);
    fclose(file);

	for(int p=0;p<nwords_per_row;p++) 
	{
		result[p] ^= result1[p];
	}
	return 0;
  
}

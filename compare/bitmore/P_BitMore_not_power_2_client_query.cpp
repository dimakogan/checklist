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
	
	constexpr uint8_t lognitems = 20;
	constexpr uint8_t nservers = 5;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t soundness = 128;
    constexpr size_t nkeys = std::ceil(std::log2(nservers)) + soundness;
	constexpr size_t item = 401;
	int err;

	//Client: query generation
    uint8_t * query;
	err = posix_memalign((void**)&query, sizeof(__m256i), (nitems) * sizeof(uint8_t));
	if(err) perror("Error in memalign for query");	
	for (int j = 0; j < nitems; ++j)
    {
        query[j] = arc4random_uniform(nservers);
    }
	for(size_t k = 0; k < nservers; ++k) 
	{
		query[item] = k;
		std::string filename = "p_bm_np2_query_to_server_" + std::to_string(k);
		FILE* file = fopen(filename.c_str(), "wb");
	    fwrite(query, nitems, sizeof(uint8_t), file);
	    fclose(file);
	}
	free(query);
	return 0;
  
}
 
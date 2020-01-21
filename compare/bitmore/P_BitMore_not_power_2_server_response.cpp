#include <bsd/stdlib.h> // for arc4random_buf
#include <cmath>        // for std::log2 and std::ceil
#include <cstring>      // for std::memcpy
#include <x86intrin.h>

#include "dpf.h"

#include <iostream>
#include <string>

#include <unistd.h>
#include <sys/mman.h>
#include <fcntl.h>

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
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = nservers-1;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	constexpr size_t soundness = 128;
    constexpr size_t nkeys = std::ceil(std::log2(nservers)) + soundness;

	int err;

	AES_KEY aeskey;
	AES_set_encrypt_key(_mm_set_epi64x(597349, 121379), &aeskey);

	typedef record<nbytes_per_word, nwords_per_row> record;
	size_t alloc_size = nitems * sizeof(record);
    
	
	//reading the database
	void * database_;
	err = posix_memalign((void**)&database_, sizeof(__m256i), alloc_size);
	if(err) perror("Error in memalign for database_");

	int file_database = open("database", O_RDONLY);
    if (file_database == -1) perror("Error opening file for reading");
    void *temp = mmap(NULL, alloc_size, PROT_READ, MAP_SHARED, file_database, 0);
    if(temp == MAP_FAILED) perror("Error in mmap");
    memcpy(database_,(char *) temp,alloc_size);
    munmap(temp,alloc_size);
    close(file_database);
	record * database = reinterpret_cast<record *>(database_);
	
	uint8_t * query;
	err = posix_memalign((void**)&query, sizeof(__m256i), nitems * sizeof(uint8_t));
	if(err) perror("Error in memalign for query");

	FILE* file2 = fopen(argv[1], "rb");
    fread (query, sizeof(uint8_t), nitems, file2);
    fclose(file2);

	//response word generation
	word<nbytes_per_word> qtemplate = { 0 };
	for (size_t i = 0; i < nitems; ++i)
	{
		if (query[i] != nwords_per_row) qtemplate ^= database[i][query[i]];
	}

	FILE* file3 = fopen(argv[2], "wb");
    fwrite(&qtemplate, 1, sizeof(qtemplate), file3);
    fclose(file3);

	free(database_);
	free(query);
	return 0;
  
}

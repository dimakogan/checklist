#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-result"

#include <cmath>        // for std::log2 and std::ceil
#include <cstring>      // for std::memcpy
#include <x86intrin.h>

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
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t item = 101;
	constexpr size_t nwords_per_row = 3;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));

	int err;

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

	word<nbytes_per_word> zerow = { 0 };
	//response from 1st server
	std::string filename = "p_chor_response_from_server_0";
	FILE* file4 = fopen(filename.c_str(), "rb");
	record result = { zerow };
    fread (&result, sizeof(result), 1, file4);
    fclose(file4);

    //response from 2nd server
    std::string filename2 = "p_chor_response_from_server_1";
	FILE* file = fopen(filename2.c_str(), "rb");
	record result1 = { zerow };
    fread (&result1, sizeof(result1), 1, file);
    fclose(file);

	for(int p=0;p<nwords_per_row;p++) 
	{
		result[p] ^= result1[p];
	}

	//test correctness
	for (size_t i = 0; i < nwords_per_row; ++i)
	{
		result[i] ^= database[item][i];
	}

	bool correct_ness = true;
	for (size_t i = 0; i < nwords_per_row; ++i)
	{
		if(result[i].m256[0][0]!=0x00) {correct_ness = false; printf("Incorrect protocol!\n");break;}
	}
	if (correct_ness) printf("Correct protocol!\n");
	free(database_);
	return 0;
  
}

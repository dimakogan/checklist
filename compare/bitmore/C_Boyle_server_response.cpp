#include <bsd/stdlib.h> // for arc4random_buf
#include <chrono>
#include <cmath>        // for std::log2 and std::ceil
#include <cstring>
#include <iostream>
#include <string>
#include <x86intrin.h>

#include "dpf.h"

#include <unistd.h>
#include <sys/mman.h>
#include <fcntl.h>

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-result"


using namespace std::chrono;
using namespace std;

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
	using namespace dpf;
	constexpr uint8_t lognitems = 20;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = 3;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	int err;
	typedef record<nbytes_per_word, nwords_per_row> record;
	size_t alloc_size = nitems * sizeof(record);

	AES_KEY aeskey;
	AES_set_encrypt_key(_mm_set_epi64x(597349, 121379), &aeskey);

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

	__m128i * s;
	uint8_t * t;
	err = posix_memalign((void**)&s, sizeof(__m256i), dpf::dpf_key<nitems>::output_length * sizeof(__m128i));
	if(err) perror("Error in memalign for s");
	t = (uint8_t *)malloc(dpf::dpf_key<nitems>::output_length * sizeof(uint8_t));

	//receive dpf_query from client
	FILE* file = fopen(argv[1], "rb");
	dpf_key<nitems> dpfkey;
    fread (&dpfkey, sizeof(dpf_key<nitems>), 1, file);
    fclose(file);

 	auto time_server_s = steady_clock::now();

    //query expansion
    dpf::evalfull(aeskey, dpfkey, s, t);
    size_t query_length = std::ceil((float)nitems / 64);
	uint64_t * expanded_query = (uint64_t*) malloc(sizeof(uint64_t) * query_length);
	for(int i=0;i<query_length/2;i++) {
		expanded_query[i*2]  = s[i][0];
		expanded_query[i*2+1] = s[i][1];
	}

	//response generation
	word<nbytes_per_word> zerow = { 0 };
	record result = { zerow };
	for (size_t k = 0; k < query_length; ++k) 
	{
		uint64_t bitset = expanded_query[k];
		while (bitset != 0) {

		  const int nextbit = __builtin_ctzll(bitset);//trailing zero, i.e. the row number where is a 1
		  if( k * 64 + nextbit >= nitems ){break;}
		  for(int p=0;p<nwords_per_row;p++) 
		  {
		    result[p] ^= database[k*64+nextbit][p];
		  }
		  bitset ^= bitset & -bitset;//vanishes LSB 1
		}
	}

	auto time_server_e = steady_clock::now();
    auto time_server_us = duration_cast<microseconds>(time_server_e - time_server_s).count();
	cout << "PIRServer reply generation time: " << (double)time_server_us / 1000 
         << " query length " << sizeof(uint64_t) * query_length
		 << endl;


	FILE* file2 = fopen(argv[2], "wb");
    fwrite(&result, 1, sizeof(result), file2);
    fclose(file2);
    
    free(s);
    free(t);
    free(expanded_query);
    free(database_);
	return 0;
  
}

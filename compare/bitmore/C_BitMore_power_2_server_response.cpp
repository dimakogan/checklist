#include "bitmore.h"
#include <iostream>
#include <string>
#include <unistd.h>
#include <sys/mman.h>
#include <fcntl.h>


int main(int argc, char * argv[])
{
	
	constexpr uint8_t lognitems = 20;
	constexpr uint8_t nservers = 4;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = nservers-1;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	constexpr size_t nkeys = std::log2(nservers);

	int err;

	AES_KEY aeskey;
	AES_set_encrypt_key(_mm_set_epi64x(597349, 121379), &aeskey);

	typedef bitmore::record<nbytes_per_word, nwords_per_row> record;
	typedef bitmore::query<nitems, nkeys, nservers> query;
	typedef bitmore::query<nitems, nkeys, nservers>::request dpf_query;
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
	
	uint8_t * expanded_query;
	err = posix_memalign((void**)&expanded_query, sizeof(__m256i), nitems * sizeof(uint8_t));
	if(err) perror("Error in memalign for expanded_query");

	__m128i ** s = (__m128i**)malloc(nkeys * sizeof(__m128i *));
	uint8_t ** t = (uint8_t**)malloc(nkeys * sizeof(uint8_t *));
	for (size_t i = 0; i < nkeys; ++i)
	{
		err = posix_memalign((void**)&s[i], sizeof(__m256i), dpf::dpf_key<nitems>::output_length * sizeof(__m128i));
		if(err) perror("Error in memalign for s");
		t[i] = (uint8_t *)malloc(dpf::dpf_key<nitems>::output_length * sizeof(uint8_t));
	}

	//Query expansion at the server
	FILE* file2 = fopen(argv[1], "rb");
	dpf_query dpf_q;
    fread (&dpf_q, sizeof(dpf_q), 1, file2);
    fclose(file2);
	bitmore::evalfull<nkeys,nitems,nservers>(aeskey, dpf_q, s, t, expanded_query);

	//response word generation
	bitmore::word<nbytes_per_word> qtemplate = { 0 };
	for (size_t i = 0; i < nitems; ++i)
	{
		if (expanded_query[i] != nwords_per_row) qtemplate ^= database[i][expanded_query[i]];
	}

	FILE* file3 = fopen(argv[2], "wb");
    fwrite(&qtemplate, 1, sizeof(qtemplate), file3);
    fclose(file3);

	for (size_t i = 0; i < nkeys; ++i)
	{
		free(s[i]);
		free(t[i]);
	}
	free(s);
	free(t);
	free(database_);
	free(expanded_query);
	return 0;
  
}
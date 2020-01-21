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
	constexpr size_t item = 201;

	int err;

	constexpr size_t nwords_per_row = nservers-1;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	constexpr size_t nkeys = std::log2(nservers);

	typedef bitmore::record<nbytes_per_word, nwords_per_row> record;
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

	//response word from s+1 th server
	std::string filename = "c_bm_p2_response_from_server_" + std::to_string(nservers-1);
	FILE* file4 = fopen(filename.c_str(), "rb");
	bitmore::word<nbytes_per_word> qqtemplate = { 0 };
    fread (&qqtemplate, sizeof(qqtemplate), 1, file4);
    fclose(file4);

    //response word from 1 to s th server
	record result = { qqtemplate };
	for (size_t k = 0; k < nservers-1; ++k)
	{	std::string filename2 = "c_bm_p2_response_from_server_" + std::to_string(k);
		FILE* file5 = fopen(filename2.c_str(), "rb");
		bitmore::word<nbytes_per_word> temp_word = { 0 };
	    fread (&temp_word, sizeof(temp_word), 1, file5);
	    fclose(file5);
		result[k] ^= temp_word;
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
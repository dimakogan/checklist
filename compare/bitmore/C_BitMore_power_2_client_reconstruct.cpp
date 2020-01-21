#include "bitmore.h"


int main(int argc, char * argv[])
{
	constexpr uint8_t nservers = 4;
	constexpr size_t nbytes_per_row = 512*2;
	constexpr size_t nwords_per_row = nservers-1;
	constexpr size_t nbytes_per_word = std::ceil(nbytes_per_row / static_cast<double>(nwords_per_row));
	constexpr size_t item = 201;
	
	typedef bitmore::record<nbytes_per_word, nwords_per_row> record;

	//response word from s+1 th server
	std::string filename = "c_bm_p2_response_from_server_" + std::to_string(nservers-1);
	FILE* file4 = fopen(filename.c_str(), "rb");
	bitmore::word<nbytes_per_word> qqtemplate = { 0 };
    fread (&qqtemplate, sizeof(qqtemplate), 1, file4);
    fclose(file4);

    //response words from remaining 1 to s th server
	record result = { qqtemplate };
	for (size_t k = 0; k < nservers-1; ++k)
	{	std::string filename2 = "c_bm_p2_response_from_server_" + std::to_string(k);
		FILE* file5 = fopen(filename2.c_str(), "rb");
		bitmore::word<nbytes_per_word> temp_word = { 0 };
	    fread (&temp_word, sizeof(temp_word), 1, file5);
	    fclose(file5);
		result[k] ^= temp_word;
	}
	return 0;
}

#include "bitmore.h"
#include <iostream>
#include <string>


int main(int argc, char * argv[])
{
	
	constexpr uint8_t lognitems = 20;
	constexpr uint8_t nservers = 4;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t nkeys = std::ceil(std::log2(nservers));
	constexpr size_t item = 201;

	typedef bitmore::query<nitems, nkeys, nservers> query;
	typedef bitmore::query<nitems, nkeys, nservers>::request dpf_query;

	//Client: query generation
	AES_KEY aeskey;
	AES_set_encrypt_key(_mm_set_epi64x(597349, 121379), &aeskey);
	query q(aeskey, item);
	dpf_query dpf_q;
	for(size_t k = 0; k < nservers; ++k) 
	{
		dpf_q = q.get_request(k);
		std::string filename = "c_bm_p2_query_to_server_" + std::to_string(k);
		FILE* file = fopen(filename.c_str(), "wb");
	    fwrite(&dpf_q, 1, sizeof(dpf_q), file);
	    fclose(file);
	}
	return 0;
  
}
 
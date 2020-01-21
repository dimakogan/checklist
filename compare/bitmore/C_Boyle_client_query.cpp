#include "dpf.h"
#include <iostream>
#include <string>
#include <cmath> 


int main(int argc, char * argv[])
{
	using namespace dpf;

	constexpr uint8_t lognitems = 20;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t item = 301;

	AES_KEY aeskey;
	AES_set_encrypt_key(_mm_set_epi64x(597349, 121379), &aeskey);

	//Client: query generation
	dpf_key<nitems> dpfkey[2];
	dpf::gen(aeskey, item, dpfkey);

	std::string filename = "c_boyle_query_to_server_0";
	FILE* file = fopen(filename.c_str(), "wb");
    fwrite(&dpfkey[0], 1, sizeof(dpf_key<nitems>), file);
    fclose(file);

	std::string filename2 = "c_boyle_query_to_server_1";
	FILE* file2 = fopen(filename2.c_str(), "wb");
    fwrite(&dpfkey[1], 1, sizeof(dpf_key<nitems>), file2);
    fclose(file2);
	return 0;
  
}
 
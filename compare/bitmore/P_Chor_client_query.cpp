#include <bsd/stdlib.h> // for arc4random_buf
#include <iostream>
#include <string>
#include <cmath> 

int main(int argc, char * argv[])
{
	constexpr uint8_t lognitems = 20;
	constexpr size_t nitems = (1ULL << lognitems);
	constexpr size_t item = 101;

	//Client: query generation
	size_t query_length = std::ceil((float)nitems / 64);
	uint64_t * query = (uint64_t*) malloc(sizeof(uint64_t) * query_length);
	arc4random_buf(query, query_length*sizeof(uint64_t));
	std::string filename = "p_chor_query_to_server_0";
	FILE* file = fopen(filename.c_str(), "wb");
    fwrite(query, query_length, sizeof(uint64_t), file);
    fclose(file);

	query[item/64] ^= 1ULL << (item%64);
	std::string filename2 = "p_chor_query_to_server_1";
	FILE* file2 = fopen(filename2.c_str(), "wb");
    fwrite(query, query_length, sizeof(uint64_t), file2);
    fclose(file2);
    free(query);
	return 0;
  
}
 
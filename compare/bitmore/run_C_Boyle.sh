#!/bin/bash
DV=1

g++ -std=c++1z -O3 C_Boyle_client_query.cpp -o C_Boyle_client_query -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_Boyle_server_response.cpp -o C_Boyle_server_response -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_Boyle_client_reconstruct.cpp -o C_Boyle_client_reconstruct -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_Boyle_client_reconstruct_with_correctness_proof.cpp -o C_Boyle_client_reconstruct_with_correctness_proof -mavx2 -march=native -lbsd


./C_Boyle_client_query
./C_Boyle_server_response c_boyle_query_to_server_0 c_boyle_response_from_server_0
./C_Boyle_server_response c_boyle_query_to_server_1 c_boyle_response_from_server_1
./C_Boyle_client_reconstruct
./C_Boyle_client_reconstruct_with_correctness_proof
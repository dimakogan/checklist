#!/bin/bash
DV=1

g++ -std=c++1z -O3 C_BitMore_power_2_client_query.cpp -o C_BitMore_power_2_client_query -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_BitMore_power_2_server_response.cpp -o C_BitMore_power_2_server_response -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_BitMore_power_2_client_reconstruct.cpp -o C_BitMore_power_2_client_reconstruct -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 C_BitMore_power_2_client_reconstruct_with_correctness_proof.cpp -o C_BitMore_power_2_client_reconstruct_with_correctness_proof -mavx2 -march=native -lbsd


./C_BitMore_power_2_client_query
./C_BitMore_power_2_server_response c_bm_p2_query_to_server_0 c_bm_p2_response_from_server_0
./C_BitMore_power_2_server_response c_bm_p2_query_to_server_1 c_bm_p2_response_from_server_1
./C_BitMore_power_2_server_response c_bm_p2_query_to_server_2 c_bm_p2_response_from_server_2
./C_BitMore_power_2_server_response c_bm_p2_query_to_server_3 c_bm_p2_response_from_server_3
./C_BitMore_power_2_client_reconstruct
./C_BitMore_power_2_client_reconstruct_with_correctness_proof
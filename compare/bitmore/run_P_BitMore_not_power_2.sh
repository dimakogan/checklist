#!/bin/bash
DV=1

g++ -std=c++1z -O3 P_BitMore_not_power_2_client_query.cpp -o P_BitMore_not_power_2_client_query -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_BitMore_not_power_2_server_response.cpp -o P_BitMore_not_power_2_server_response -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_BitMore_not_power_2_client_reconstruct.cpp -o P_BitMore_not_power_2_client_reconstruct -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_BitMore_not_power_2_client_reconstruct_with_correctness_proof.cpp -o P_BitMore_not_power_2_client_reconstruct_with_correctness_proof -mavx2 -march=native -lbsd


./P_BitMore_not_power_2_client_query
./P_BitMore_not_power_2_server_response p_bm_np2_query_to_server_0 p_bm_np2_response_from_server_0
./P_BitMore_not_power_2_server_response p_bm_np2_query_to_server_1 p_bm_np2_response_from_server_1
./P_BitMore_not_power_2_server_response p_bm_np2_query_to_server_2 p_bm_np2_response_from_server_2
./P_BitMore_not_power_2_server_response p_bm_np2_query_to_server_3 p_bm_np2_response_from_server_3
./P_BitMore_not_power_2_server_response p_bm_np2_query_to_server_4 p_bm_np2_response_from_server_4
./P_BitMore_not_power_2_client_reconstruct
./P_BitMore_not_power_2_client_reconstruct_with_correctness_proof
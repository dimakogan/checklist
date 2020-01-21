#!/bin/bash
DV=1

g++ -std=c++1z -O3 P_Chor_client_query.cpp -o P_Chor_client_query -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_Chor_server_response.cpp -o P_Chor_server_response -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_Chor_client_reconstruct.cpp -o P_Chor_client_reconstruct -mavx2 -march=native -lbsd
g++ -std=c++1z -O3 P_Chor_client_reconstruct_with_correctness_proof.cpp -o P_Chor_client_reconstruct_with_correctness_proof -mavx2 -march=native -lbsd


./P_Chor_client_query
./P_Chor_server_response p_chor_query_to_server_0 p_chor_response_from_server_0
./P_Chor_server_response p_chor_query_to_server_1 p_chor_response_from_server_1
./P_Chor_client_reconstruct
./P_Chor_client_reconstruct_with_correctness_proof
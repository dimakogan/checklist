#!/bin/bash
DV=1

CFLAGS=-std=c++1z -O3 -mavx2 -march=native -lbsd

g++ $CFLAGS C_Boyle_client_query.cpp -o C_Boyle_client_query 
g++ $CFLAGS C_Boyle_server_response.cpp -o C_Boyle_server_response 
g++ $CFLAGS C_Boyle_client_reconstruct.cpp -o C_Boyle_client_reconstruct 
g++ $CFLAGS -O3 C_Boyle_client_reconstruct_with_correctness_proof.cpp -o C_Boyle_client_reconstruct_with_correctness_proof 


./C_Boyle_client_query
./C_Boyle_server_response c_boyle_query_to_server_0 c_boyle_response_from_server_0
./C_Boyle_server_response c_boyle_query_to_server_1 c_boyle_response_from_server_1
./C_Boyle_client_reconstruct
./C_Boyle_client_reconstruct_with_correctness_proof

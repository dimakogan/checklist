#include <iostream>
#include "../xor.h"

int main()
{
    int row_len = 31;
    uint8_t db[row_len * 7];
    long long unsigned int idx1 = 3;
    long long unsigned int idx2 = 6;
    for (int i = 0; i < row_len; i++)
    {
        db[idx1*row_len + i] = 'X';
        db[idx2*row_len + i] = 'X' ^ (uint8_t)i;
    }

    long long unsigned int elems[] = {idx1*row_len, idx2*row_len};
    uint8_t out[row_len+3];
    xor_rows(db, sizeof(db), elems, 2, row_len, out+1);

    for (int i = 0; i < row_len; i++)
    {
        std::cout << int(out[1+i]) << " ";
    }
    std::cout << std::endl;
}
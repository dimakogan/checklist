#include <chrono>
#include <iostream>
#include <random>
#include <string>
#include <vector>

using namespace std;


string random_record(int rec_size_bytes) {
    default_random_engine generator;
    uniform_int_distribution<int> distribution(0,255);
    string s;
     for (int i = 0; i < rec_size_bytes; ++i) {
        s += (char)(distribution(generator));
    }
    return s;
}

void sequential_read(const vector<string>& db, int num_reads) {
    string item;
    char sum = 0;
    for (auto i = 0; i < num_reads; ++i) {
        item = db[i % db.size()];
        // Touch some character of the string to avoid optimizations.
        sum += item[2];
    }
    cerr << "Finished " << num_reads << " reads " << sum << endl;
}


int main(int argc, const char* argv[]) {
    enum {
        PROGRAM = 0,
        RECORD_SIZE_IN_BYTES = 1,
        NUM_RECORDS = 2,
        NUM_READS = 3,
        NUM_ARGS = 4
    };

    if (argc < NUM_ARGS) {
        cerr << "Usage: " 
            << argv[PROGRAM] 
            << " <RECORD_SIZE_IN_BYTES> <NUM_RECORDS> <NUM_READS>" << endl;
        return 1;
    }
    int rec_size_bytes = stoi(argv[RECORD_SIZE_IN_BYTES]);
    int num_recs = stoi(argv[NUM_RECORDS]);
    int num_reads = stoi(argv[NUM_READS]);

    cerr << "Reading " << num_reads << " records "
        << "from a database of " << num_recs 
        << " records of size " << rec_size_bytes << " bytes each." << endl;
    vector<string> db;
    for (auto i = 0; i < num_recs; ++i) {
        db.push_back(random_record(rec_size_bytes));
    }
    cerr << "Finished setting up DB" << endl;
    auto start = chrono::steady_clock::now();
    sequential_read(db, num_reads);
    auto end = chrono::steady_clock::now();
    int duration_ms = (end - start).count() / (1000 * 1000);
    cout << "Measured duration of reads: " << duration_ms << endl;
    cerr << "Throughput: " << ((double)num_reads * rec_size_bytes * 1000 / duration_ms) 
        << " Bytes/sec" << endl;
}
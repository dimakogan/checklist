Step 1) create a 1 GB database, keeping name 'database', by the command: dd if=/dev/urandom of=database bs=1M count=1024

Step 2) Run the script to execute the complete "computationally private BitMore protocol with power-of-2 servers" by the command ./run_C_BitMore_power_2.sh

Step 3) Run the script to execute the complete "computationally private BitMore protocol with not-power-of-2 servers" by the command ./run_C_BitMore_not_power_2.sh

Step 4) Run the script to execute the complete "perfectly private Chor et al. protocol" by the command ./run_P_Chor.sh

Step 5) Run the script to execute the complete "computationally private Boyle et al. protocol" by the command ./run_C_Boyle.sh

Step 6) Run the script to execute the complete "computationally private Boyle et al. protocol" by the command ./run_P_BitMore_not_power_2.sh
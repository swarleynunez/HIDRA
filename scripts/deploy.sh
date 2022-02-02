#!/usr/bin/env bash

# Clean previous runs
COMMAND="cd /home/pi/.ethereum/bin/; \
pkill hidra; \
docker stop \$(docker ps -a -q) && docker rm \$(docker ps -a -q); \
rm hidra hidra.out hidra.err"
sshpass -p "pi_node0_eth" ssh pi@192.168.43.10 "$COMMAND"
sshpass -p "pi_node1_eth" ssh pi@192.168.43.11 "$COMMAND"
sshpass -p "pi_node2_eth" ssh pi@192.168.43.12 "$COMMAND"

# Compile HIDRA and send executable and .env file to each node
env GOOS=linux GOARCH=arm GOARM=7 go build -o bin/hidra main.go
sshpass -p "pi_node0_eth" scp bin/hidra config/.env pi@192.168.43.10:/home/pi/.ethereum/bin
sshpass -p "pi_node1_eth" scp bin/hidra config/.env pi@192.168.43.11:/home/pi/.ethereum/bin
sshpass -p "pi_node2_eth" scp bin/hidra config/.env pi@192.168.43.12:/home/pi/.ethereum/bin

# Execute HIDRA in each node logging output and errors
COMMAND="cd /home/pi/.ethereum/bin/; \
nohup ./hidra run > hidra.out 2> hidra.err &"
sshpass -p "pi_node0_eth" ssh pi@192.168.43.10 "$COMMAND"

COMMAND="cd /home/pi/.ethereum/bin/; \
nohup ./hidra run > hidra.out 2> hidra.err &"
sshpass -p "pi_node1_eth" ssh pi@192.168.43.11 "$COMMAND"

COMMAND="cd /home/pi/.ethereum/bin/; \
nohup ./hidra run > hidra.out 2> hidra.err &"
sshpass -p "pi_node2_eth" ssh pi@192.168.43.12 "$COMMAND"

## Get the controller SC address
#COMMAND="cd /home/pi/.ethereum/bin/; \
#nohup ./superfog > superfog.out 2> superfog.err & \
#sleep 10; \
#sed -n '/CONTROLLER_ADDR=/p' .env"
#ADDR=$(sshpass -p "pi_node0_eth" ssh pi@192.168.43.10 "$COMMAND")
#
## Set new controller SC address
#COMMAND="cd /home/pi/.ethereum/bin/; \
#sed -i '1s/^/$ADDR\n/' .env; \
#nohup ./superfog > superfog.out 2> superfog.err &"
#sshpass -p "pi_node1_eth" ssh pi@192.168.43.11 "$COMMAND"
#sshpass -p "pi_node2_eth" ssh pi@192.168.43.12 "$COMMAND"
#sed -i '/CONTROLLER_ADDR/d' .env; \

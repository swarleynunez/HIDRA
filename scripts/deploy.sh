#!/usr/bin/env bash

# Clean previous executions
COMMAND="cd /home/pi/.ethereum/bin/; \
pkill superfog; \
docker stop \$(docker ps -a -q) && docker rm \$(docker ps -a -q); \
sed -i '/CONTROLLER_ADDR/d' .env; \
rm superfog superfog.out"
sshpass -p "pi_node0_eth" ssh pi@172.24.100.10 "$COMMAND"
sshpass -p "pi_node1_eth" ssh pi@172.24.100.11 "$COMMAND"
sshpass -p "pi_node2_eth" ssh pi@172.24.100.45 "$COMMAND"
#sshpass -p "pi_node3_eth" ssh pi@172.24.100.48 "$COMMAND"
#sshpass -p "pi_node4_eth" ssh pi@172.24.100.77 "$COMMAND"

# Compile superfog client and send it to each node
env GOOS=linux GOARCH=arm GOARM=7 go build -o superfog cmd/main.go
sshpass -p "pi_node0_eth" scp superfog pi@172.24.100.10:/home/pi/.ethereum/bin
sshpass -p "pi_node1_eth" scp superfog pi@172.24.100.11:/home/pi/.ethereum/bin
sshpass -p "pi_node2_eth" scp superfog pi@172.24.100.45:/home/pi/.ethereum/bin
#sshpass -p "pi_node3_eth" scp superfog pi@172.24.100.48:/home/pi/.ethereum/bin
#sshpass -p "pi_node4_eth" scp superfog pi@172.24.100.77:/home/pi/.ethereum/bin

# Get the controller SC address
COMMAND="cd /home/pi/.ethereum/bin/; \
nohup ./superfog > superfog.out & \
sleep 5; \
sed -n '/CONTROLLER_ADDR=/p' .env"
ADDR=$(sshpass -p "pi_node0_eth" ssh pi@172.24.100.10 "$COMMAND")

# Set new controller SC address
COMMAND="cd /home/pi/.ethereum/bin/; \
sed -i '1s/^/$ADDR\n/' .env; \
nohup ./superfog > superfog.out &"
sshpass -p "pi_node1_eth" ssh pi@172.24.100.11 "$COMMAND"
sshpass -p "pi_node2_eth" ssh pi@172.24.100.45 "$COMMAND"
#sshpass -p "pi_node3_eth" ssh pi@172.24.100.48 "$COMMAND"
#sshpass -p "pi_node4_eth" ssh pi@172.24.100.77 "$COMMAND"

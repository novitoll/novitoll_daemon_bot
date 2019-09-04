#!/bin/bash

# Script to check the local development env

items=(docker-compose docker go)

for item in ${items[@]};do
	if ! hash "$item" 2>/dev/null; then
		echo "[-] Please install $item."
		exit 1
	else
		echo "[+] OK. $item"
	fi
done

if [ "$(pwd)" != "$GOPATH" ];then
	echo "[-] \$GOPATH differs. Please set a proper \$GOPATH."
	exit 1
fi

echo "[+] Done."

#!/bin/bash
trap exit SIGINT
while true
do
	kubectl port-forward -n percona-everest deployment/percona-everest 8080:8080
done

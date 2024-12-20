#!/bin/bash
set -euo pipefail

# Collect the STUN endpoint domain and port
if [[ $# -ne 2 ]]; then
    echo "usage: $0 <domain> <port>"
    echo ""
    echo "for example: rbmk sh $0 stun.l.google.com 19302"
    exit 1
fi
DOMAIN=$1
PORT=$2

# Create the working directory in which to save results
WORKDIR=$(rbmk timestamp)-stunlookup
rbmk mkdir -p $WORKDIR

# Define the name of the files containing A and AAAA lookup results
A_JSONL=$WORKDIR/a.jsonl
AAAA_JSONL=$WORKDIR/aaaa.jsonl

# Define the named pipe to pass addresses around
ADDR_PIPE=$WORKDIR/addrs.pipe

# Run the DNS lookups in the background and publish results using the named pipe
(rbmk dig --measure --logs $A_JSONL +short=ip $DOMAIN A | rbmk pipe write $ADDR_PIPE) &
(rbmk dig --measure --logs $AAAA_JSONL +short=ip $DOMAIN AAAA | rbmk pipe write $ADDR_PIPE) &

# Read unique resolved addresses, make endpoints, and measure using STUN
# printing out the discovered reflexive IP addresses
COUNT=0
rbmk pipe read --writers 2 $ADDR_PIPE | rbmk ipuniq --port $PORT | while read EPNT; do
	STUN_JSONL=$WORKDIR/stun${COUNT}.jsonl
	rbmk stun --measure --logs $STUN_JSONL $EPNT | rbmk ipuniq -E
	COUNT=$((COUNT + 1))
done

# Compress the measurements into a tarball
TARBALL=$WORKDIR.tar.gz
rbmk tar -czf $TARBALL $WORKDIR
rbmk rm -rf $WORKDIR

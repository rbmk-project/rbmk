#doc:
#doc: rbmk_stun_lookup [flags]
#doc:
#doc: Performs a STUN transaction returning the public
#doc: IP address(es) of the current host.
#doc:
#doc: Flags
#doc:
#doc: --stun-hostname string (default "stun.l.google.com")
#doc:     The hostname of the STUN server to use.
#doc:
#doc: --stun-port string (default "19302")
#doc:     The port of the STUN server to use.
#doc:
#doc: --stun-addr-ipv4 string (default "")
#doc:     Well-known IPv4 address of the STUN server.
#doc:
#doc: --stun-addr-ipv6 string (default "")
#doc:     Well-known IPv6 address of the STUN server.
#doc:
#doc: --dns-protocol string (default "udp")
#doc:     The DNS protocol to use for resolving the STUN server hostname.
#doc:
#doc: --dns-server string (default "8.8.8.8")
#doc:     The DNS server to use for resolving the STUN server hostname.
#doc:
#doc: Files
#doc:
#doc: dig4.jsonl
#doc:     The JSONL log file for the IPv4 DNS resolution.
#doc:
#doc: dig4.err
#doc:     The error log file for the IPv4 DNS resolution.
#doc:
#doc: dig6.jsonl
#doc:     The JSONL log file for the IPv6 DNS resolution.
#doc:
#doc: dig6.err
#doc:     The error log file for the IPv6 DNS resolution.
#doc:
#doc: stun*.jsonl
#doc:     The JSONL log files for the STUN transactions.
#doc:
#doc: addrs.txt
#doc:     The list of public IP addresses discovered.
#doc:
#doc: All files will be written in the current working directory.
#doc:
rbmk_stun_lookup() {
	# Configure the default values.
	local stun_hostname="stun.l.google.com"
	local stun_port="19302"
	local stun_addr_ipv4=""
	local stun_addr_ipv6=""
	local dns_protocol="udp"
	local dns_server="8.8.8.8"

	# Process function arguments to override defaults.
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--stun-hostname)
			stun_hostname="$2"
			shift 2
			;;
		--stun-port)
			stun_port="$2"
			shift 2
			;;
		--stun-addr-ipv4)
			stun_addr_ipv4="$2"
			shift 2
			;;
		--stun-addr-ipv6)
			stun_addr_ipv6="$2"
			shift 2
			;;
		--dns-protocol)
			dns_protocol="$2"
			shift 2
			;;
		--dns-server)
			dns_server="$2"
			shift 2
			;;
		*)
			echo "rbmk_stun_lookup: unknown argument: $1" >&2
			return 1
			;;
		esac
	done

	# Resolve IPv4 in the background publishing the resolved
	# addresses to the pipe named `a.pipe`. Using a pipe allows
	# to stream IP addresses as they are resolved, which is
	# useful when you use `--dns-protocol udp=wait-duplicates`
	# to detect GFW like DNS poisoning with duplicates.
	#
	# Note: we're using a short pipe name to avoid Unix domain
	# socket issues (path names longer than ~90 chars aren't
	# portable and cause the socket bind to fail).
	(rbmk dig --measure \
		--logs dig4.jsonl \
		+short=ip \
		"+${dns_protocol}" \
		"@${dns_server}" \
		"${stun_hostname}" \
		A \
		2>dig4.err | rbmk pipe write a.pipe) &

	# Same as above but for IPv6.
	(rbmk dig --measure \
		--logs dig6.jsonl \
		+short=ip \
		"+${dns_protocol}" \
		"@${dns_server}" \
		"${stun_hostname}" \
		AAAA \
		2>dig6.err | rbmk pipe write a.pipe) &

	# Include well-known static IP addresses if available
	# to circumvent potential DNS based censorship.
	#
	# We need to count the number of workers streaming IP
	# addresses to pass its valud to `rbmk pipe read`.
	local workers=2
	if [[ -n ${stun_addr_ipv4} ]]; then
		(echo "${stun_addr_ipv4}" | rbmk pipe write a.pipe) &
		workers="$((workers + 1))"
	fi
	if [[ -n ${stun_addr_ipv6} ]]; then
		(echo "${stun_addr_ipv6}" | rbmk pipe write a.pipe) &
		workers="$((workers + 1))"
	fi

	# Collect the STUN endpoints and perform the STUN transactions
	# as soon as the IP addresses are resolved. The STUN transactions
	# will be performed sequentially since they write into the same
	# output file. The `ipuniq` command is used to filter out duplicate
	# IP addresses returned by DNS or STUN servers.
	local count="0"
	rbmk pipe read --writers "${workers}" a.pipe |
		rbmk ipuniq -p "${stun_port}" |
		while read endpoint; do
			rbmk stun --measure \
				--max-time 5 \
				--logs "stun${count}.jsonl" \
				"${endpoint}" \
				2>"stun${count}.err" | rbmk ipuniq -E >>addrs.txt
			count="$((count + 1))"
		done

	# Wait for background processes to finish
	wait
}

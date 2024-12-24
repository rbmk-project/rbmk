#doc:
#doc: rbmk_stun_lookup [flags]
#doc:
#doc: Performs a STUN transaction returning the public
#doc: IP address(es) of the current host.
#doc:
#doc: Flags
#doc:
#doc: --disable-ipv6
#doc:     Disables IPv6 resolution and measurement.
#doc:
#doc: --dns-protocol string (default "udp")
#doc:     The DNS protocol to use for resolving the STUN server hostname.
#doc:
#doc: --dns-server string (default "8.8.8.8")
#doc:     The DNS server to use for resolving the STUN server hostname.
#doc:
#doc: --start-user-options
#doc:     Indicates that user-provided options start here. Allows to
#doc:     ignore options that would conflict with the generated configuration.
#doc:
#doc: --stop-user-options
#doc:     Indicates that user-provided options stop here.
#doc:
#doc: --stun-addr-ipv4 string (default "")
#doc:     Well-known IPv4 address of the STUN server.
#doc:
#doc: --stun-addr-ipv6 string (default "")
#doc:     Well-known IPv6 address of the STUN server.
#doc:
#doc: --stun-hostname string (default "stun.l.google.com")
#doc:     The hostname of the STUN server to use.
#doc:
#doc: --stun-port string (default "19302")
#doc:     The port of the STUN server to use.
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
#doc: addrs*.txt
#doc:     The list of public IP addresses discovered by a transaction.
#doc:
#doc: All files will be written in the current working directory.
#doc:
rbmk_stun_lookup() {
	# Configure the default values.
	local disable_ipv6="0"
	local dns_protocol="udp"
	local dns_server="8.8.8.8"
	local stun_addr_ipv4=""
	local stun_addr_ipv6=""
	local stun_hostname="stun.l.google.com"
	local stun_port="19302"
	local user_options="0"

	# Process function arguments to override defaults.
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--disable-ipv6)
			disable_ipv6="1"
			shift 1
			;;
		--dns-protocol)
			dns_protocol="$2"
			shift 2
			;;
		--dns-server)
			dns_server="$2"
			shift 2
			;;
		--start-user-options)
			user_options=1
			shift
			;;
		--stop-user-options)
			user_options=0
			shift
			;;
		--stun-addr-ipv4)
			if [[ $user_options -eq 0 ]]; then
				stun_addr_ipv4="$2"
			fi
			shift 2
			;;
		--stun-addr-ipv6)
			if [[ $user_options -eq 0 ]]; then
				stun_addr_ipv6="$2"
			fi
			shift 2
			;;
		--stun-hostname)
			if [[ $user_options -eq 0 ]]; then
				stun_hostname="$2"
			fi
			shift 2
			;;
		--stun-port)
			if [[ $user_options -eq 0 ]]; then
				stun_port="$2"
			fi
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

	# We need to count the number of workers streaming IP
	# addresses to pass its value to `rbmk pipe read`.
	local workers=1

	# Same as above but for IPv6.
	if [[ $disable_ipv6 -eq "0" ]]; then
		(rbmk dig --measure \
			--logs dig6.jsonl \
			+short=ip \
			"+${dns_protocol}" \
			"@${dns_server}" \
			"${stun_hostname}" \
			AAAA \
			2>dig6.err | rbmk pipe write a.pipe) &
		workers="$((workers + 1))"
	fi

	# Include well-known static IP addresses if available
	# to circumvent potential DNS based censorship.
	if [[ -n "${stun_addr_ipv4}" ]]; then
		(echo "${stun_addr_ipv4}" | rbmk pipe write a.pipe) &
		workers="$((workers + 1))"
	fi
	if [[ -n "${stun_addr_ipv6}" && "${disable_ipv6}" -eq "0" ]]; then
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
			fmt_count="$(printf "%08d" "${count}")"
			rbmk stun --measure \
				--max-time 5 \
				--logs "stun${fmt_count}.jsonl" \
				"${endpoint}" \
				2>"stun${fmt_count}.err" |
				rbmk ipuniq -E >>"addrs${fmt_count}.txt"
			count="$((count + 1))"
		done

	# Wait for background processes to finish
	wait
}

#doc: prints help message for the STUN lookup script.
rbmk_stun_lookup_help() {
	rbmk cat <<\EOF | rbmk markdown

# STUN servers measurement script

## Usage

```
rbmk sh SCRIPT [flags]
```

## Description

Measures a list of STUN servers for reachability by performing
STUN binding transactions with a list of STUN servers.

Uses the following RBMK commands:

1. `rbmk dig` to resolve the STUN server's hostname to IP addresses.

2. `rbmk stun` to perform STUN binding transactions.

This script is generated by `rbmk generate stun_lookup`.

## Flags

### `--disable-ipv6`

Disables IPv6 resolution and measurement.

### `--dns-protocol PROTO`

Selects the DNS protocol `rbmk dig` should use:

- `udp`: DNS-over-UDP (the default).

- `udp=wait-duplicates`: DNS-over-UDP and wait until timeout
for duplicate responses, which may indicate DNS poisoning.

- `tcp`: DNS-over-TCP.

- `tls`: DNS-over-TLS.

- `https`: DNS-over-HTTPS.

See `rbmk dig --help` for more information.

### `--dns-server SERVER`

Selects the DNS `SERVER` to use. Here you could use an IP address
or a domain name. When using a domain name, we will use the system
resolver to map it to IP addresses and use the first IP address
that works. The default is to use the `8.8.8.8` server.

See `rbmk dig --help` for more information.

## Files

This script creates a tarball archive containing the measurements
and named `YYYYMMDDTHHMMSSZ-stunlookup.tar.gz`.

The archive contains a directory for each measured STUN server. Each
directory contains the following files:

- `dig4.jsonl`: Structured logs file for the IPv4 DNS resolution.

- `dig4.err`: Error log file for the IPv4 DNS resolution.

- `dig6.jsonl`: Structured logs file for the IPv6 DNS resolution.

- `dig6.err`: Error log file for the IPv6 DNS resolution.

- `stun${fmt_count}.jsonl`: Structured logs file for a STUN transaction.

- `stun${fmt_count}.err`: Error log file for a STUN transaction.

- `addrs${fmt_count}.txt`: IP addresses discovered by a STUN transaction.

The `${fmt_count}` variable is a zero-padded sequential number.

## Exit Status

Returns `0` on success, `1` on failure.

EOF
}

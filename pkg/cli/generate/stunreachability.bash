#doc: rbmk_measure_stun_reachability DESTDIR DOMAIN PORT
#doc:
#doc: Performs STUN lookups for each resolved IP address of DOMAIN
#doc: using the given PORT. Creates the DESTDIR to store:
#doc:
#doc: DNS resolution outputs:
#doc:
#doc: - a-logs.json: IPv4 resolution logs and results
#doc: - a-stdout.txt, a-stderr.txt: IPv4 resolution output
#doc: - aaaa-logs.json: IPv6 resolution logs and results
#doc: - aaaa-stdout.txt, aaaa-stderr.txt: IPv6 resolution output
#doc:
#doc: STUN measurement outputs (per resolved IP):
#doc:
#doc: - stun-NNNN.jsonl: STUN measurement logs
#doc: - stun-NNNN-stdout.txt: STUN reflexive endpoint
#doc: - stun-NNNN-stderr.txt: STUN error messages
#doc:
#doc: Where NNNN is a zero-padded sequential number.
rbmk_measure_stun_reachability() {
	# Gather command line arguments
	local destdir="$1"
	shift
	local domain="$1"
	shift
	local port="$1"
	shift

	# Create the destination directory
	rbmk mkdir -p "$destdir"

	# Resolve the domain name A records in the background
	local a_jsonl="$destdir/a-logs.jsonl"
	local a_stdout="$destdir/a-stdout.txt"
	local a_stderr="$destdir/a-stderr.txt"
	rbmk dig --measure --logs "$a_jsonl" +short=ip "$domain" A 1>"$a_stdout" 2>"$a_stderr"

	# Resolve the domain name AAAA records in the background
	local aaaa_jsonl="$destdir/aaaa-logs.jsonl"
	local aaaa_stdout="$destdir/aaaa-stdout.txt"
	local aaaa_stderr="$destdir/aaaa-stderr.txt"
	rbmk dig --measure --logs "$aaaa_jsonl" +short=ip "$domain" AAAA 1>"$aaaa_stdout" 2>"$aaaa_stderr"

	# Obtain unique endpoints
	local endpoints="$(rbmk ipuniq --port "$port" "$a_stdout" "$aaaa_stdout")"

	# Measure STUN reachability for each endpoint
	local count=0
	for epnt in "$endpoints"; do
		local formatted_count="$(printf "%02d" "$count")"
		local stun_jsonl="$destdir/stun-${formatted_count}.jsonl"
		local stun_stdout="$destdir/stun-${formatted_count}-stdout.txt"
		local stun_stderr="$destdir/stun-${formatted_count}-stderr.txt"
		rbmk stun --max-time 5 --measure --logs "$stun_jsonl" "$epnt" 1>"$stun_stdout" 2>"$stun_stderr"
		count=$((count + 1))
	done
}

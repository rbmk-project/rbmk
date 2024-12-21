#!/bin/bash
set -euo pipefail

#doc: rbmk_stun_lookup DESTDIR ADDR PORT
#doc:
#doc: Creates the DESTDIR directory, if needed, and runs
#doc: a STUN lookup using the given IP ADDR and PORT.
#doc:
#doc: Writes the STUN logs to DESTDIR/logs.jsonl, the
#doc: command stdout to DESTDIR/stdout.txt, and the stderr
#doc: to DESTDIR/stderr.txt.
rbmk_stun_lookup() {
	local destdir=$1
	shift

	local addr=$1
	shift

	local port=$1
	shift

	rbmk mkdir -p "$destdir"

	rbmk stun \
		--max-time 5 \
		--measure \
		--logs "$destdir/logs.jsonl" \
		$(echo $addr | rbmk ipuniq -p $port) \
		1>"$destdir/stdout.txt" \
		2>"$destdir/stderr.txt"
}

#doc: rbmk_measure_sni_blocking DESTDIR ADDR PORT SNI
#doc:
#doc: Creates the DESTDIR directory, if needed, and runs
#doc: a netcat measurement using the given IP ADDR and PORT
#doc: with the SNI extension set to SNI.
#doc:
#doc: Writes the netcat logs to DESTDIR/logs.jsonl, the
#doc: command stdout to DESTDIR/stdout.txt and the command
#doc: stderr to DESTDIR/stderr.txt.
rbmk_measure_sni_blocking() {
	local destdir=$1
	shift

	local addr=$1
	shift

	local port=$1
	shift

	local sni=$1
	shift

	rbmk mkdir -p "$destdir"

	rbmk nc \
		--measure \
		-zcw5 \
		--alpn h2 \
		--alpn http/1.1 \
		--sni "$sni" \
		--logs "$destdir/logs.jsonl" \
		"$addr" "$port" \
		1>"$destdir/stdout.txt" \
		2>"$destdir/stderr.txt"
}

#doc: rbmk_ui_print_progress CURRENT TOTAL
#doc:
#doc: Prints a progress bar to the standard output.
rbmk_ui_print_progress() {
    if [[ "${RBMK_TRACE:-0}" = "1" ]]; then
        return
    fi

	local current=$1
	shift

	local total=$1
	shift

	local width=40
	local progress=$((current * width / total))
	local filled=""
	local empty=""

	for ((idx = 0; idx < progress; idx++)); do
		filled="$filled#"
	done
	for ((idx = progress; idx < width; idx++)); do
		empty="$empty-"
	done

	printf '\r[%s%s] %d%% (%d/%d)' "$filled" "$empty" $((current * 100 / total)) "$current" "$total"

	if [[ $current -ge $total ]]; then
		printf '\n'
	fi
}

#doc: rbmk_output_format_dir_prefix COUNT
#doc:
#doc: Formats the COUNT as an 8-digit zero-padded string.
rbmk_output_format_dir_prefix() {
	local count=$1
	shift
	printf "%08d" $count
}

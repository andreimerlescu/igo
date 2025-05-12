#!/bin/bash

# Function to parse command line arguments
parse_arguments() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
			--help|-h|help)
				print_usage
				exit 0
				;;
			--*)
				key="${1/--/}" # Remove '--' prefix
				key="${key//-/_}" # Replace '-' with '_' to match params[key]
				if [[ -n "${2}" && "${2:0:1}" != "-" ]]; then
					params[$key]="$2"
					shift 2
				else
					echo "Error: Missing value for $1" >&2
					exit 1
				fi
				;;
			*)
				echo "Unknown option: $1" >&2
				print_usage
				exit 1
				;;
		esac
	done
}

# print_usage
function print_usage(){
	echo "Usage: ${0} [OPTIONS]"
	mapfile -t sorted_keys < <(for param in "${!params[@]}"; do echo "$param"; done | sort)
	local -i padSize=3;
	for param in "${sorted_keys[@]}"; do local -i len="${#param}"; (( len > padSize )) && padSize=len; done
	((padSize+=3)) # add right buffer
	for param in "${sorted_keys[@]}"; do
		local d; local p; p="${params[$param]}"; { [[ -n "${p}" ]] && [[ "${#p}" != 0 ]] && d=" (default = '${p}')"; } || d=""
		echo "       --$(pad "$padSize" "${param}") ${documentation[$param]}${d}"
	done
}

function pad() {
  printf "%-${1}s\n" "${2}";
}


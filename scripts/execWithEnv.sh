#!/bin/bash

# Enable some useful safety options
set -o nounset  # Treat unset variables as an error
set -o errexit  # Exit on any command failure (but we'll handle specific exits manually)

echo "## $0 received NUM ARGS: $#"

ENV_FILENAME='.env'

if [[ $# -eq 1 ]]; then
  BIN_FILENAME="$1"
elif [[ $# -eq 2 ]]; then
  BIN_FILENAME="$1"
  ENV_FILENAME="${2:-.env}"  # Fixed: correct default value syntax
else
  echo "## 💥💥 Usage: $0 <executable-path> [env-file]"
  echo "##       First argument: path to the executable binary"
  echo "##       Second argument (optional): .env file (defaults to '.env')"
  exit 1
fi

echo "## Will try to run: ${BIN_FILENAME}"
echo "## With environment from: ${ENV_FILENAME}"

# Check if the .env file exists and is readable
if [[ ! -r "$ENV_FILENAME" ]]; then
  echo "## 💥💥 Error: Environment file '${ENV_FILENAME}' not found or not readable"
  exit 1
fi

# Check if the binary exists and is executable
if [[ ! -x "$BIN_FILENAME" ]]; then
  echo "## 💥💥 Error: '${BIN_FILENAME}' is not an executable file or does not have execute permission"
  exit 1
fi

echo "## Executing ${BIN_FILENAME} with variables from ${ENV_FILENAME}..."

# Load environment variables safely (still has limitations — see note below)
set -a  # Automatically export all variables
# shellcheck disable=SC1090
source <(sed -e '/^#/d' -e '/^\s*$/d' -e "s/'/'\\\''/g" -e "s/=\(.*\)/='\1'/g" "$ENV_FILENAME")
set +a

# Finally run the binary
exec "$BIN_FILENAME"  # Use exec to replace the shell process (cleaner)

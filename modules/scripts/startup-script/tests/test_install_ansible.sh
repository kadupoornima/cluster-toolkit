#!/bin/bash
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This test verifies that install_ansible.sh handles existing files/directories correctly.

set -e

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
INSTALL_SCRIPT="$SCRIPT_DIR/../files/install_ansible.sh"
TEST_DIR=$(mktemp -d)

# Cleanup on exit
trap 'rm -rf "$TEST_DIR"' EXIT

echo "Running test in $TEST_DIR"

# Mock environment
mkdir -p "$TEST_DIR/usr/bin"
mkdir -p "$TEST_DIR/etc"
mkdir -p "$TEST_DIR/usr/local"
mkdir -p "$TEST_DIR/var/lib/apt/lists"

# Create a copy of the script with modified paths
TEST_SCRIPT="$TEST_DIR/install_ansible_test.sh"
cp "$INSTALL_SCRIPT" "$TEST_SCRIPT"

# sed replacement to redirect paths to TEST_DIR
# Note: We use | as delimiter
sed -i "s|/usr/bin|$TEST_DIR/usr/bin|g" "$TEST_SCRIPT"
sed -i "s|/etc|$TEST_DIR/etc|g" "$TEST_SCRIPT"
sed -i "s|/var/lib/apt|$TEST_DIR/var/lib/apt|g" "$TEST_SCRIPT"
sed -i "s|/usr/local|$TEST_DIR/usr/local|g" "$TEST_SCRIPT"
# We also need to fix the case where install_python3_dnf/apt calls `command -v python3` which returns absolute path.
# But we are mocking the environment for `ln -s` mainly.

chmod +x "$TEST_SCRIPT"

# Mock python3 availability if needed, but we rely on system python being available.
# The script calls `command -v python3` to find python. It should find the system one.
# But then it installs venv in TEST_DIR.

echo "Test Case 1: Running script for the first time..."
"$TEST_SCRIPT"

if [ ! -d "$TEST_DIR/etc/ansible" ]; then
    echo "Error: /etc/ansible not created"
    exit 1
fi

if [ ! -L "$TEST_DIR/usr/bin/ansible" ]; then
    echo "Error: /usr/bin/ansible symlink not created"
    exit 1
fi

echo "Test Case 1 Passed."

echo "Test Case 2: Running script again (idempotency/re-run)..."
# The script should not fail if /etc/ansible exists
# The script should not fail if symlinks exist
"$TEST_SCRIPT"
echo "Test Case 2 Passed."

echo "Test Case 3: Running script with existing conflicting files..."
# Create a dummy file where a symlink should be
rm "$TEST_DIR/usr/bin/ansible"
touch "$TEST_DIR/usr/bin/ansible"

# Remove ansible.cfg but keep directory to trigger mkdir failure if -p is missing
rm "$TEST_DIR/etc/ansible/ansible.cfg"
# Directory $TEST_DIR/etc/ansible still exists

"$TEST_SCRIPT"

if [ ! -L "$TEST_DIR/usr/bin/ansible" ]; then
    echo "Error: /usr/bin/ansible was not replaced with a symlink"
    exit 1
fi

if [ ! -f "$TEST_DIR/etc/ansible/ansible.cfg" ]; then
    echo "Error: ansible.cfg was not recreated"
    exit 1
fi

echo "Test Case 3 Passed."

echo "All tests passed."

#!/bin/sh

# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# set -x;

# Env variable `TEST_PLUGIN_EXIT_STATUS` is used by testing.
status=${TEST_PLUGIN_EXIT_STATUS:-0};

echo "Running" $(basename $0) "(path: $0) with status($status)...";
max=10
for i in $(bash -c "echo {1..${max}}"); do
    echo $i;
    sleep $i;
done

echo "Displaying Plugin Manager (PM) Config file path: ${PM_CONF_FILE}"

if [ ${status} -eq 0 ]; then
    echo "Done($status)!";
else
    echo "Fail($status)";
fi

exit ${status};

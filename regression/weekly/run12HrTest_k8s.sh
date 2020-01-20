#!/bin/bash -e
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#


######################################################################
### Run one group of the tests in weekly test suite in k8s

CurrentDirectory=$(cd `dirname $0` && pwd)
FabricTestDir="$(echo $CurrentDirectory | awk -F'/fabric-test/' '{print $1}')/fabric-test"
WEEKLYDIR="$FabricTestDir/regression/weekly"
export GinkoTests=true
echo "========== Performance PTE 12Hr test =========="
cd $WEEKLYDIR && ginkgo
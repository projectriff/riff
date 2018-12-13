#!/bin/bash

# TODO make FATS channel tests modular enough to use here
test_name=echo

kail --ns $NAMESPACE > $test_name.logs &
kail_test_pid=$!

kail --ns knative-serving --ns knative-eventing > $test_name.controller.logs &
kail_controller_pid=$!

riff service create correlator --image projectriff/correlator:fats --namespace $NAMESPACE
wait_kservice_ready correlator $NAMESPACE

riff channel create $test_name --cluster-provisioner in-memory-channel --namespace $NAMESPACE
riff subscription create $test_name --channel $test_name --subscriber correlator --namespace $NAMESPACE

wait_channel_ready $test_name $NAMESPACE
wait_subscription_ready $test_name $NAMESPACE
sleep 5

input_data=riff
riff service invoke correlator /$NAMESPACE/$test_name --namespace $NAMESPACE --text -- \
  -H "knative-blocking-request: true" \
  -w'\n' \
  -d $input_data | tee $test_name.out
expected_data=riff
actual_data=`cat $test_name.out | tail -1`

kill $kail_test_pid $kail_controller_pid
riff subscription delete $test_name --namespace $NAMESPACE
riff channel delete $test_name --namespace $NAMESPACE
riff service delete correlator --namespace $NAMESPACE

if [[ "$actual_data" != "$expected_data" ]]; then
  fats_echo "Test Logs:"
  cat $test_name.logs
  echo ""
  fats_echo "Controller Logs:"
  cat $test_name.controller.logs
  echo ""
  fats_echo "${ANSI_RED}Test did not produce expected result${ANSI_RESET}";
  echo "   expected data: $expected_data"
  echo "   actual data: $actual_data"
  exit 1
fi

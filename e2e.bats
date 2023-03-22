#!/usr/bin/env bats

@test "reject because too much memory is requested" {
  run kwctl run \
    --allow-context-aware \
    -r test_data/ns_not_valid.json \
    --replay-host-capabilities-interactions test_data/session.yml \
    annotated-policy.wasm

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request rejected
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*false') -ne 0 ]
  [ $(expr "$output" : ".*LimitsMemory limit.*") -ne 0 ]
}

@test "accept because amount of resources requested is not exceeding resource quota" {
  run kwctl run \
    --allow-context-aware \
    -r test_data/ns_valid.json \
    --replay-host-capabilities-interactions test_data/session.yml \
    annotated-policy.wasm

  # this prints the output when one the checks below fails
  echo "output = ${output}"

  # request accepted
  [ "$status" -eq 0 ]
  [ $(expr "$output" : '.*allowed.*true') -ne 0 ]
}

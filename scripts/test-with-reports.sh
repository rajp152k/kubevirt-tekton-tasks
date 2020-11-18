#!/usr/bin/env bash

set -o pipefail

RET_CODE=0

ARTIFACT_DIR="${ARTIFACT_DIR:=dist}"
ARTIFACT_DIR="$(readlink -m "${ARTIFACT_DIR}")"
TEST_OUT="${ARTIFACT_DIR}/test.out"
COVER_OUT="${ARTIFACT_DIR}/coverage.out"
JUNIT_XML="${ARTIFACT_DIR}/junit.xml"
COVERAGE_HTML="${ARTIFACT_DIR}/coverage.html"
FAKE_GOPATH_ROOT="/tmp/goroot-kubevirt-tekton-tasks"
FAKE_KV_GOPATH="${FAKE_GOPATH_ROOT}/src/github.com/kubevirt"
FAKE_REPO_GOPATH="${FAKE_KV_GOPATH}/kubevirt-tekton-tasks"

rm -rf "${TEST_OUT}" "${COVER_OUT}" "${JUNIT_XML}" "${COVERAGE_HTML}" "${FAKE_GOPATH_ROOT}"
mkdir -p "${ARTIFACT_DIR}"

pushd modules > /dev/null || exit 1
  export DIST_DIR=dist

  for MODULE_DIR in $(ls | grep -vE "^(tests)$"); do
    pushd "$MODULE_DIR" > /dev/null || continue
      make test-verbose
      CURRENT_RET_CODE=$?
      if [ "${CURRENT_RET_CODE}" -ne 0 ]; then
        RET_CODE=${CURRENT_RET_CODE}
      fi
      cat ${DIST_DIR}/test.out >> "${TEST_OUT}"

      if [ -f "${COVER_OUT}" ]; then
        sed "/^mode.*/d" dist/coverage.out >> "${COVER_OUT}" # remove first line with mode
      else
        cp ${DIST_DIR}/coverage.out "${COVER_OUT}"
      fi

    popd > /dev/null || continue
  done
popd > /dev/null || exit 1

if type go-junit-report > /dev/null; then
  go-junit-report < "${TEST_OUT}" > "${JUNIT_XML}"
fi

mkdir -p "${FAKE_KV_GOPATH}"
ln -s "$(pwd)" "${FAKE_REPO_GOPATH}"
GOPATH="${FAKE_GOPATH_ROOT}"
export "GOPATH=${GOPATH}"

go tool cover -html "${COVER_OUT}" -o "${COVERAGE_HTML}"

rm -rf "${FAKE_GOPATH_ROOT}"

exit $RET_CODE
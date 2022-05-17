#!/bin/bash

set -e

workdir=$(dirname $(realpath $0))
version=$(cat version 2>/dev/null)
bin_name="polaris-cleanup"
if [ "${GOOS}" == "" ]; then
    GOOS=$(go env GOOS)
fi
if [ "${GOARCH}" == "" ]; then
    GOARCH=$(go env GOARCH)
fi
folder_name="polaris-cleanup-release_${version}.${GOOS}.${GOARCH}"
pkg_name="${folder_name}.zip"
if [ "${GOOS}" == "windows" ]; then
    bin_name="polaris-cleanup.exe"
fi
echo "GOOS is ${GOOS}, binary name is ${bin_name}"

cd $workdir

# 清理环境
rm -rf ${folder_name}
rm -f "${pkg_name}"

# 编译
rm -f ${bin_name}

# 禁止 CGO_ENABLED 参数打开
export CGO_ENABLED=0

build_date=$(date "+%Y%m%d.%H%M%S")
package="github.com/polarismesh/polaris-cleanup/common"
go build -o ${bin_name} -ldflags="-X ${package}.Version=${version} -X ${package}.BuildDate=${build_date}"

# 打包
mkdir -p ${folder_name}
cp ${bin_name} ${folder_name}
cp polaris-cleanup.yaml ${folder_name}
cp -r tool ${folder_name}/
zip -r "${pkg_name}" ${folder_name}
#md5sum ${pkg_name} > "${pkg_name}.md5sum"

if [[ $(uname -a | grep "Darwin" | wc -l) -eq 1 ]]; then
    md5 ${pkg_name} >"${pkg_name}.md5sum"
else
    md5sum ${pkg_name} >"${pkg_name}.md5sum"
fi

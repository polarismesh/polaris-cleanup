#!/bin/bash

workdir=$(dirname $(realpath $0))
version=$(cat version 2>/dev/null)
folder_name="polaris-server-agent_${version}"
pkg_name="${folder_name}.tar.gz"
bin_name="polaris-server-agent"

cd $workdir

# 清理环境
rm -rf ${folder_name}
rm -f "${pkg_name}"

# 编译
rm -f ${bin_name}


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o ${bin_name}

# 打包
mkdir -p ${folder_name}
mv ${bin_name} ${folder_name}
tar -czvf "${pkg_name}" ${folder_name}
md5sum ${pkg_name} > "${pkg_name}.md5sum"

/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

var (
	Version   string
	BuildDate string
)

// GetVersion 获取版本号
func GetVersion() string {
	if Version == "" {
		return "v1.0.0"
	}

	return Version
}

// GetRevision 获取完整版本号信息，包括时间戳的
func GetRevision() string {
	if Version == "" || BuildDate == "" {
		return "v1.0.0"
	}

	return Version + "." + BuildDate
}

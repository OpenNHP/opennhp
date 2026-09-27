//go:build linux

package ebpf

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// getBootTimeNanos 返回单调时钟的纳秒时间戳，用于构造 eBPF map 中的
// expire_time。该时钟必须与 BPF helper bpf_ktime_get_ns() 所在的时钟域
// 一致（内核侧为 CLOCK_MONOTONIC，而非开机时间时钟）——若此处使用
// 开机时间时钟，宿主机挂起/唤醒后用户态写入的 expire_time 会发生漂移，
// 再次出现"敲门放行超时未关闭"的症状。
//
// 函数名沿用历史命名 getBootTimeNanos 以避免改动所有调用点；
// 实现已切换到单调时钟。
func getBootTimeNanos() (uint64, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
		return 0, fmt.Errorf("clock_gettime failed: %v", err)
	}
	return uint64(ts.Sec)*1e9 + uint64(ts.Nsec), nil
}

// +build !linux

package hostobject

func FileFdCollect() (map[string]int64, error) {
	info := make(map[string]int64)
	info["allocated"] = -1
	info["maximum"] = -1
	return info, nil
}

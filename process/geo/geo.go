package geo

import (
	"fmt"
	"path/filepath"

	ipL "github.com/ip2location/ip2location-go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	Db           *ipL.DB
	Ip2LocDbPath = filepath.Join(datakit.InstallDir, "data", "iploc.bin")
)

func Geo(ip string) (ipL.IP2Locationrecord, error) {
	if Db == nil {
		return ipL.IP2Locationrecord{}, fmt.Errorf("ip2location db nil")
	}
	return Db.Get_all(ip)
}

func init() {
	var err error
	Db, err = ipL.OpenDB(Ip2LocDbPath)
	if err != nil {
		fmt.Printf("Open %v db err %v", Ip2LocDbPath, err)
	}
}

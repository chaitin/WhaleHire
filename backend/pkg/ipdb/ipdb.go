package ipdb

import (
	"embed"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"

	"github.com/chaitin/WhaleHire/backend/domain"
)

//go:embed ip2region.xdb
var ipdbFiles embed.FS

type IPDB struct {
	searcher *xdb.Searcher
	logger   *slog.Logger
}

func NewIPDB(logger *slog.Logger) (*IPDB, error) {
	cBuff, err := xdb.LoadContentFromFS(ipdbFiles, "ip2region.xdb")
	if err != nil {
		return nil, fmt.Errorf("load xdb index failed: %w", err)
	}
	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		return nil, fmt.Errorf("new xdb reader failed: %w", err)
	}
	return &IPDB{searcher: searcher, logger: logger.With("module", "ipdb")}, nil
}

func (a *IPDB) Lookup(ip string) (*domain.IPAddress, error) {
	// 输入验证
	if ip == "" {
		return nil, fmt.Errorf("empty ip address")
	}

	// 处理IPv6地址，ip2region库主要支持IPv4
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		a.logger.Warn("invalid ip address format", "ip", ip)
		return &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}, nil
	}

	// 如果是IPv6地址，使用默认值
	if parsedIP.To4() == nil {
		a.logger.Debug("ipv6 address detected, using default values", "ip", ip)
		return &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}, nil
	}

	// 检查是否为内网IP
	if isPrivateIP(parsedIP) {
		a.logger.Debug("private ip address detected", "ip", ip)
		return &domain.IPAddress{
			IP:       ip,
			Country:  "中国",
			Province: "内网",
			City:     "内网",
		}, nil
	}

	// 使用 recover 捕获 panic，防止 slice bounds 错误导致服务崩溃
	var region string
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				a.logger.Error("panic occurred during ip lookup", "ip", ip, "error", r)
				err = fmt.Errorf("ip lookup panic: %v", r)
			}
		}()
		region, err = a.searcher.SearchByStr(ip)
	}()

	if err != nil {
		a.logger.Warn("search ip failed, using default values", "ip", ip, "error", err)
		return &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}, nil
	}

	ipInfo := strings.Split(region, "|")
	if len(ipInfo) != 5 {
		a.logger.Warn("invalid ip info format, using default values", "ip", ip, "region", region)
		return &domain.IPAddress{
			IP:       ip,
			Country:  "未知",
			Province: "未知",
			City:     "未知",
		}, nil
	}

	country := ipInfo[0]
	province := ipInfo[2]
	city := ipInfo[3]
	if country == "0" || country == "" {
		country = "未知"
	}
	if province == "0" || province == "" {
		province = "未知"
	}
	if city == "0" || city == "" {
		city = "未知"
	}

	return &domain.IPAddress{
		IP:       ip,
		Country:  country,
		Province: province,
		City:     city,
	}, nil
}

// isPrivateIP 检查是否为内网IP地址
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// 检查私有网络地址段
	privateBlocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},     // 10.0.0.0/8
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},  // 172.16.0.0/12
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)}, // 192.168.0.0/16
	}

	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

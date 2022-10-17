package uid

import (
	"bytes"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
)

type UIDRepo interface {
	GenID() (string, error)
	TraceID() string
	TradeID() string
}

var uidOnce sync.Once
var onceUID *UID

type Option func(u *UID)

type UIDOptions struct{}

func (UIDOptions) WithConf(conf config.Config) Option {
	return func(m *UID) {
		m.conf = conf
	}
}

func (UIDOptions) WithStartTime(date string) Option {
	return func(m *UID) {
		m.startTime = date
	}
}

func (UIDOptions) WithMachineID(mId int) Option {
	return func(m *UID) {
		m.machineID = mId
	}
}

type UID struct {
	conf      config.Config
	startTime string
	machineID int
	sf        *sonyflake.Sonyflake
	currDate  atomic.Value
}

var timeFunc = func() string {
	return time.Now().Format("20060102")
}

func NewUIDRepo(options ...Option) UIDRepo {
	uidOnce.Do(func() {
		onceUID = new(UID)

		for _, option := range options {
			option(onceUID)
		}

		if onceUID.conf == nil {
			onceUID.conf = config.NewNullConfig()
		}

		if onceUID.machineID == 0 {
			onceUID.machineID = onceUID.conf.GetInt("uid_machine_id")
		}

		if onceUID.machineID == 0 {
			ipId, _ := onceUID.lower16BitPrivateIP()
			onceUID.machineID = ipId + os.Getpid()%128
		}

		if onceUID.startTime == "" {
			onceUID.startTime = onceUID.conf.GetString("uid_start_time")
		}

		if onceUID.startTime == "" {
			onceUID.startTime = timeFunc()[:4] + "0101"
		}

		onceUID.currDate.Store(timeFunc())

		//初始化sn
		t, _ := time.Parse("20060102", onceUID.startTime)
		settings := sonyflake.Settings{
			StartTime: t,
			MachineID: func() (uint16, error) {
				return uint16(onceUID.machineID), nil
			},
		}
		onceUID.sf = sonyflake.NewSonyflake(settings)
		go func() {
			t := time.NewTicker(3 * time.Second)
			for {
				select {
				case <-t.C:
					onceUID.currDate.Store(timeFunc())
				}
			}
		}()
	})
	return onceUID
}

func (u *UID) lower16BitPrivateIP() (int, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return int(ip[2])<<8 + int(ip[3]), nil
}

func (u *UID) GenID() (string, error) {
	id, err := u.sf.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

func (u *UID) date() string {
	return u.currDate.Load().(string)
}

func (u *UID) TradeID() string {
	var buf bytes.Buffer
	buf.WriteString(u.date())
	id, _ := u.GenID()
	buf.WriteString(id)
	return buf.String()
}

func (u *UID) TraceID() string {
	var buf bytes.Buffer
	buf.WriteString(u.date())
	buf.WriteString(ksuid.New().String())
	return buf.String()
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

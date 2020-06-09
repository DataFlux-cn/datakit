package replication

import (
	"context"
	"fmt"
	"log"

	// "log"
	"sync"
	"sync/atomic"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/jackc/pgx"
)

type stream struct {
	rep *Replication
	// 订阅配置
	sub *Subscribe
	// 当前 wal 位置
	receivedWal uint64
	flushWal    uint64
	// 复制连接
	replicationConn *pgx.ReplicationConn
	// 取消
	cancel context.CancelFunc
	// ack 锁
	sendStatusLock sync.Mutex

	points []*influxdb.Point
}

func (s *stream) getReceivedWal() uint64 {
	return atomic.LoadUint64(&s.receivedWal)
}

func (s *stream) setReceivedWal(val uint64) {
	atomic.StoreUint64(&s.receivedWal, val)
}

func (s *stream) getFlushWal() uint64 {
	return atomic.LoadUint64(&s.flushWal)
}

func (s *stream) getStatus() (*pgx.StandbyStatus, error) {
	return pgx.NewStandbyStatus(s.getReceivedWal(), s.getFlushWal(), s.getFlushWal())
}

func newStream(sub *Subscribe) *stream {
	return &stream{sub: sub}
}

func (s *stream) start(r *Replication, ctx context.Context, wg *sync.WaitGroup) error {
	s.rep = r
	s.points = []*influxdb.Point{}

	// log.SetFlags(log.Lshortfile | log.LstdFlags | log.LUTC)
	defer wg.Done()

	configUpdatePoint(s.sub)
	// log.Printf("start stream for %s\n", s.sub.SlotName)
	ctx, s.cancel = context.WithCancel(ctx)

	config := pgx.ConnConfig{
		Host:     s.sub.Host,
		Port:     s.sub.Port,
		Database: s.sub.Database,
		User:     s.sub.User,
		Password: s.sub.Password,
	}

	conn, err := pgx.ReplicationConnect(config)
	if err != nil {
		log.Printf("E! [Replication] subscribe %s:%d, create replication connect failed: %s\n", s.sub.Host, s.sub.Port, err.Error())
		return err
	}

	s.replicationConn = conn

	slotname := s.sub.SlotName

	_, _, err = conn.CreateReplicationSlotEx(slotname, "test_decoding")
	if err != nil {
		// 42710 means replication slot already exists
		if pgerr, ok := err.(pgx.PgError); !ok || pgerr.Code != "42710" {
			log.Printf("E! [Replication] subscribe %s:%d, create replication slot err: %s\n", s.sub.Host, s.sub.Port, err.Error())
			return err
		}
	}

	_ = s.sendStatus()

	if err := conn.StartReplication(slotname, 0, -1); err != nil {
		log.Printf("E! [Replication] subscribe %s:%d, start connection err: %s\n", s.sub.Host, s.sub.Port, err.Error())
		return err
	}

	log.Printf("I! [Replication] subscribe %s:%d start\n", s.sub.Host, s.sub.Port)

	return s.runloop(ctx)
}

func (s *stream) stop() error {
	s.cancel()
	log.Printf("I! [Replication] subscribe %s:%d stop\n", s.sub.Host, s.sub.Port)
	return s.replicationConn.Close()
}

func (s *stream) runloop(ctx context.Context) error {
	defer func() {
		if err := s.stop(); err != nil {
			log.Printf("I! [Replication] subscribe %s:%d connection close failed: %s\n", s.sub.Host, s.sub.Port, err.Error())
		}
	}()

	go func() {
		for {
			if _, ok := <-time.After(5 * time.Second); ok {
				_ = s.sendStatus()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// nil
		}

		msg, err := s.replicationConn.WaitForReplicationMessage(ctx)
		if err != nil {
			if err == ctx.Err() {
				log.Printf("W! [Replication] subscribe %s:%d, wait message err: %s\n", s.sub.Host, s.sub.Port, err.Error())
				return err
			}
			if err := s.checkAndResetConn(); err != nil {
				log.Printf("W! [Replication] subscribe %s:%d, reset replication connection err: %s\n", s.sub.Host, s.sub.Port, err.Error())
			}
			continue
		}

		if msg == nil {
			log.Printf("W! [Replication] subscribe %s:%d, handle replication msg is empty\n", s.sub.Host, s.sub.Port)
			continue
		}

		if err := s.replicationMsgHandle(msg); err != nil {
			log.Printf("E! [Replication] subscribe %s:%d, handle replication msg err: %s\n", s.sub.Host, s.sub.Port, err.Error())
			continue
		}
	}
}

func (s *stream) checkAndResetConn() error {
	if s.replicationConn != nil && s.replicationConn.IsAlive() {
		return nil
	}

	time.Sleep(time.Second * 10)

	config := pgx.ConnConfig{
		Host:     s.sub.Host,
		Port:     s.sub.Port,
		Database: s.sub.Database,
		User:     s.sub.User,
		Password: s.sub.Password,
	}
	conn, err := pgx.ReplicationConnect(config)
	if err != nil {
		return err
	}

	if _, _, err := conn.CreateReplicationSlotEx(s.sub.SlotName, "test_decoding"); err != nil {
		if pgerr, ok := err.(pgx.PgError); !ok || pgerr.Code != "42710" {
			return fmt.Errorf("failed to create replication slot: %s", err)
		}
	}

	if err := conn.StartReplication(s.sub.SlotName, 0, -1); err != nil {
		_ = conn.Close()
		return err
	}

	s.replicationConn = conn

	return nil
}

// ReplicationMsgHandle handle replication msg
func (s *stream) replicationMsgHandle(msg *pgx.ReplicationMessage) error {

	// 回复心跳
	if msg.ServerHeartbeat != nil {

		if msg.ServerHeartbeat.ServerWalEnd > s.getReceivedWal() {
			s.setReceivedWal(msg.ServerHeartbeat.ServerWalEnd)
		}
		if msg.ServerHeartbeat.ReplyRequested == 1 {
			_ = s.sendStatus()
		}
	}

	if msg.WalMessage != nil {

		point, err := parse(s.sub, msg.WalMessage)
		if err != nil {
			return fmt.Errorf("invalid pgoutput msg: %s", err)
		}

		if point != nil {
			s.points = append(s.points, point)
		}
	}

	return s.flush()
}

func (s *stream) flush() (err error) {
	err = s.rep.ProcessPts(s.points)
	s.points = nil
	return err
}

// 发送心跳
func (s *stream) sendStatus() error {
	s.sendStatusLock.Lock()
	defer s.sendStatusLock.Unlock()

	// log.Printf("send heartbeat\n")
	status, err := s.getStatus()
	if err != nil {
		return err
	}
	return s.replicationConn.SendStandbyStatus(status)
}

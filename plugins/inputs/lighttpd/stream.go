package lighttpd

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type stream struct {
	lighttpd *Lighttpd
	//
	sub *Subscribe
	//
	statusURL string
	// v1 or v2
	statusVersion Version
	// mate data
	points []*influxdb.Point
}

func newStream(sub *Subscribe, lt *Lighttpd) *stream {
	var url string
	var v Version

	switch sub.LighttpdVersion {
	case "v1":
		url = fmt.Sprintf("%s?json", sub.LighttpdURL)
		v = v1
	case "v2":
		url = fmt.Sprintf("%s?format=plain", sub.LighttpdURL)
		v = v2
	default:
		// nil
	}

	return &stream{
		lighttpd:      lt,
		sub:           sub,
		statusURL:     url,
		statusVersion: v,
	}
}

func (s *stream) start(wg *sync.WaitGroup) error {
	defer wg.Done()

	if s.sub.Measurement == "" {
		err := errors.New("invalid measurement")
		log.Printf("E! [Lighttpd] subscribe '%s', err: %s\n", s.sub.LighttpdURL, err.Error())
		return err
	}

	ticker := time.NewTicker(time.Second * s.sub.Cycle)
	defer ticker.Stop()

	log.Printf("I! [Lighttpd] subscribe '%s' start\n", s.sub.LighttpdURL)

	for {
		select {
		case <-s.lighttpd.ctx.Done():
			log.Printf("I! [Lighttpd] subscribe '%s' stop\n", s.sub.LighttpdURL)
			return nil
		case <-ticker.C:
			if err := s.exec(); err != nil {
				log.Printf("E! [Lighttpd] subscribe '%s', exec err: %s\n", s.sub.LighttpdURL, err.Error())
			}
		default:
			// nil
		}
	}
}

func (s *stream) exec() error {

	pt, err := LighttpdStatusParse(s.statusURL, s.statusVersion, s.sub.Measurement)
	if err != nil {
		return err
	}

	s.points = []*influxdb.Point{pt}

	return s.flush()
}

func (s *stream) flush() (err error) {
	// fmt.Printf("%v\n", s.points)
	err = s.lighttpd.ProcessPts(s.points)
	s.points = nil

	return err
}

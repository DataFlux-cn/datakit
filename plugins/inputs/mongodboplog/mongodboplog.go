package mongodboplog

import (
	"reflect"
	"strconv"
	"time"

	"github.com/vinllen/mgo"
	"github.com/vinllen/mgo/bson"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "mongodb_oplog"

	defaultMeasurement = "mongodb_oplog"

	sampleCfg = `
# [inputs.mongodb_oplog]
# 	# MongoDB URL: mongodb://user:password@host:port/database
# 	mongodb_url="mongodb://127.0.0.1:27017"
# 	
# 	# database
# 	database="testdb"
# 	
# 	# collection
# 	collection="testdb"
# 	
# 	# tags path，is may be null
# 	tagList=[
# 		"/path",
# 		"/a/b/c/e"
# 	]
# 	
# 	# fields path, cannot empty
# 	# type in [int, float, bool, string]
# 	# example:
# 	[inputs.mongodb_oplog.fieldList]
# 		"/a/c/d" = "int"
# 		"/a/c/f[1]/e/f" = "int"
# 		# "/a/c/f\\[0\\]" = "int"
# 	
# 	# [inputs.mongodb_oplog.tags]
# 	# tags1 = "tags1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Mongodboplog{}
	})
}

type Mongodboplog struct {
	MongodbURL  string            `toml:"mongodb_url"`
	Database    string            `toml:"database"`
	Collection  string            `toml:"collection"`
	Measurement string            `toml:"measurement"`
	TagList     []string          `toml:"tagList"`
	FieldList   map[string]string `toml:"fieldList"`
	Tags        map[string]string `toml:"tags"`

	// mongodb namespace is 'database.collection'
	namespace string
	// stream iterator
	iter *mgo.Iter
	// type
	pointlist map[string]string
}

func (_ *Mongodboplog) Catalog() string {
	return "mongodb"
}

func (_ *Mongodboplog) SampleConfig() string {
	return sampleCfg
}

func (m *Mongodboplog) Run() {
	l = logger.SLogger(inputName)

	m.namespace = m.Database + "." + m.Collection
	m.pointlist = make(map[string]string)
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}

	for _, v := range m.TagList {
		m.pointlist[v] = "tags"
	}
	for k, v := range m.FieldList {
		m.pointlist[k] = v
	}

	var session *mgo.Session
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		default:
			// nil
		}

		session, err = mgo.Dial(m.MongodbURL)
		if err != nil {
			l.Errorf("failed to connect, err: %s", err.Error())
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	session.SetPoolLimit(2)
	session.SetSocketTimeout(0)
	session.SetMode(mgo.Primary, true)

	var query = bson.M{}

	// bson.MongoTimestamp int64
	// |----------32---------|-----------32-----------|
	// |   timestamp second  |          count         |
	query["ts"] = bson.M{"$gt": bson.MongoTimestamp(time.Now().Unix() << 32)}

	m.iter = session.DB("local").C("oplog.rs").Find(query).LogReplay().Tail(-1)
	defer m.iter.Close()

	l.Info("connect success")

	m.runloop()
}

func (m *Mongodboplog) runloop() {
	var lograw *bson.Raw

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		default:
			lograw = new(bson.Raw)
			for m.iter.Next(lograw) {
				p := new(PartialLog)
				if err := bson.Unmarshal(lograw.Data, p); err != nil {
					l.Error(err)
					continue
				}

				if p.Namespace != m.namespace {
					continue
				}

				switch p.Operation {
				case "i", "u", "d":
					mdata := newMgodata(m.pointlist)
					mdata.setTime(p.Timestamp)
					mdata.parse(p.Object, "/")

					data, err := mdata.makeMetric(m.Tags)
					if err != nil {
						l.Error(err)
						continue
					}
					if err := io.Feed(data, io.Metric); err != nil {
						l.Error(err)
						continue
					}
					l.Debugf("feed %d bytes to io ok", len(data))

				default:
					// nil

				}
			}
		}
	}
}

type mgodata struct {
	tags      map[string]string
	fields    map[string]interface{}
	pointlist map[string]string
	time      time.Time
}

func newMgodata(pointlist map[string]string) *mgodata {
	return &mgodata{
		tags:      make(map[string]string),
		fields:    make(map[string]interface{}),
		pointlist: pointlist,
		time:      time.Now(),
	}
}

func (md *mgodata) makeMetric(tags map[string]string) ([]byte, error) {
	for k, v := range tags {
		md.tags[k] = v
	}
	return io.MakeMetric(defaultMeasurement, md.tags, md.fields, md.time)
}

func (md *mgodata) setTime(ts bson.MongoTimestamp) {
	md.time = time.Unix(int64(ts)>>32, 0)
}

func (md *mgodata) parse(docelem bson.D, succkey string) {

	for _, elem := range docelem {
		completeKey := succkey + elem.Name

		if reflect.DeepEqual(reflect.TypeOf(elem.Value), reflect.TypeOf(bson.D{})) {
			md.parse(elem.Value.(bson.D), completeKey+"/")
		}

		if array, ok := elem.Value.([]interface{}); ok {
			for k, v := range array {
				arraykey := completeKey + "[" + strconv.Itoa(k) + "]"

				if reflect.DeepEqual(reflect.TypeOf(v), reflect.TypeOf(bson.D{})) {
					md.parse(v.(bson.D), arraykey+"/")
				}
				md.typeAssert(arraykey, v)
			}
		}

		md.typeAssert(completeKey, elem.Value)
	}
}

func (md *mgodata) typeAssert(completeKey string, value interface{}) {
	if typee, ok := md.pointlist[completeKey]; ok {
		switch typee {
		case "tags":
			if v, ok := value.(string); ok {
				md.tags[completeKey] = v
			}
		case "int":
			if v, ok := value.(float64); ok {
				md.fields[completeKey] = int64(v)
			}
		case "float":
			if v, ok := value.(float64); ok {
				md.fields[completeKey] = v
			}
		case "bool":
			if v, ok := value.(bool); ok {
				md.fields[completeKey] = v
			}
		case "string":
			if v, ok := value.(string); ok {
				md.fields[completeKey] = v
			}
		default:
			// nil
		}
	}
}

type PartialLog struct {
	Timestamp     bson.MongoTimestamp `bson:"ts"`
	Operation     string              `bson:"op"`
	Gid           string              `bson:"g"`
	Namespace     string              `bson:"ns"`
	Object        bson.D              `bson:"o"`
	Query         bson.M              `bson:"o2"`
	UniqueIndexes bson.M              `bson:"uk"`
	Lsid          interface{}         `bson:"lsid"`        // mark the session id, used in transaction
	FromMigrate   bool                `bson:"fromMigrate"` // move chunk

	/*
	 * Every field subsequent declared is NEVER persistent or
	 * transfer on network connection. They only be parsed from
	 * respective logic
	 */
	UniqueIndexesUpdates bson.M // generate by CollisionMatrix
	RawSize              int    // generate by Decorator
	SourceId             int    // generate by Validator
}

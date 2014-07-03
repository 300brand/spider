package queue

import (
	"github.com/300brand/logger"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type mongoItem struct {
	Id    bson.ObjectId
	Queue string
	Value string `bson:"_id"`
}

type mongoQueue struct {
	Name    string
	session *mgo.Session
	shard   bool
}

var _ Queue = new(mongoQueue)

func NewMongo(url string, shard bool) (q Queue, err error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return
	}
	mq := &mongoQueue{
		session: s,
		shard:   shard,
	}
	s.DB("").C("queue").EnsureIndexKey("id")
	if shard {
		err := s.DB("admin").Run(bson.M{
			"shardCollection":  s.DB("").Name + ".queue",
			"key":              bson.M{"queue": "hashed"},
			"numInitialChunks": 16,
		}, nil)
		if err != nil {
			logger.Error.Printf("sh.shardCollection(%s.queue): %s", s.DB("").Name, err)
		}
	}
	q = mq.New("default")
	return
}

func (q *mongoQueue) New(name string) Queue {
	newQueue := &mongoQueue{
		Name:    name,
		session: q.session.Copy(),
		shard:   q.shard,
	}
	return newQueue
}

func (q *mongoQueue) Dequeue() (s string, err error) {
	ch := mgo.Change{
		Remove: true,
	}
	var result mongoItem
	_, err = q.session.DB("").C("queue").Find(bson.M{"queue": q.Name}).Sort("id").Apply(ch, &result)
	if err == mgo.ErrNotFound {
		err = ErrEmpty
	}
	s = result.Value
	return
}

func (q *mongoQueue) Enqueue(s string) (err error) {
	err = q.session.DB("").C("queue").Insert(mongoItem{bson.NewObjectId(), q.Name, s})
	if mgo.IsDup(err) {
		return ErrExists
	}
	return
}

func (q *mongoQueue) Len() int {
	n, _ := q.session.DB("").C("queue").Find(bson.M{"queue": q.Name}).Count()
	return n
}

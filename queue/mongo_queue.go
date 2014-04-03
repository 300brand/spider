package queue

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type mongoItem struct {
	Id    bson.ObjectId
	Value string `bson:"_id"`
}

type mongoQueue struct {
	session *mgo.Session
	Name    string
}

var _ Queue = new(mongoQueue)

func NewMongo(url string) (q Queue, err error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return
	}
	mq := &mongoQueue{
		session: s,
	}
	q = mq.New("default")
	return
}

func (q *mongoQueue) New(name string) Queue {
	newQueue := &mongoQueue{
		Name:    name,
		session: q.session.Copy(),
	}
	newQueue.session.DB("").C(newQueue.cName()).EnsureIndexKey("id")
	return newQueue
}

func (q *mongoQueue) Dequeue() (s string, err error) {
	ch := mgo.Change{
		Remove: true,
	}
	var result mongoItem
	_, err = q.session.DB("").C(q.cName()).Find(nil).Sort("id").Apply(ch, &result)
	if err == mgo.ErrNotFound {
		err = ErrEmpty
	}
	s = result.Value
	return
}

func (q *mongoQueue) Enqueue(s string) (err error) {
	err = q.session.DB("").C(q.cName()).Insert(mongoItem{bson.NewObjectId(), s})
	if mgo.IsDup(err) {
		return ErrExists
	}
	return
}

func (q *mongoQueue) Len() int {
	n, _ := q.session.DB("").C(q.cName()).Count()
	return n
}

func (q *mongoQueue) cName() string {
	return "queue_" + strings.Replace(q.Name, ".", "_", -1)
}

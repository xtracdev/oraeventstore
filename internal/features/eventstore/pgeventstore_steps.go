package eventstore

import (
	. "github.com/lsegal/gucumber"
	"github.com/stretchr/testify/assert"
	"github.com/xtraclabs/goes"
	. "github.com/xtraclabs/goessample/testagg"
	"github.com/xtraclabs/oraeventstore"
	log "github.com/Sirupsen/logrus"
)

func init() {
	var testAgg *TestAgg
	var anotherAgg *TestAgg

	var eventStore *oraeventstore.OraEventStore
	var concurrentMax *int
	var events []goes.Event

	Given(`^a new aggregate instance$`, func() {
		log.Info("create event store")
		eventStore, _ = oraeventstore.NewPGEventStore("esusr", "password", "xe.oracle.docker", "localhost", "1521")
		if assert.NotNil(T, eventStore) {
			var err error
			testAgg, err = NewTestAgg("new foo", "new bar", "new baz")
			assert.Nil(T, err)
			assert.NotNil(T, testAgg)
		}
	})

	When(`^we check the max version in the event store$`, func() {
	})

	Then(`^the max version is 0$`, func() {
		if eventStore != nil {
			max, err := eventStore.GetMaxVersionForAggregate(testAgg.ID)
			if err != nil {
				log.Infof("Error reading max version for agg: %s", err.Error())
			}
			assert.Nil(T, err)
			if max != nil {
				assert.Equal(T, 0, *max)
			}
		}
	})

	When(`^we get the max version from the event store$`, func() {
		var err error
		concurrentMax, err = eventStore.GetMaxVersionForAggregate(testAgg.ID)
		assert.Nil(T, err)
	})

	And(`^the max version is greater than the aggregate version$`, func() {
		testAgg.Version = *concurrentMax - 1
	})

	Then(`^a concurrency error is return on aggregate store$`, func() {
		err := testAgg.Store(eventStore)
		assert.NotNil(T, err)
		assert.Equal(T, oraeventstore.ErrConcurrency, err)
	})

	Given(`^a persisted aggregate$`, func() {
		var err error
		log.Println("create an aggregate")
		anotherAgg, err = NewTestAgg("foo2", "bar2", "baz2")
		assert.Nil(T, err)
		anotherAgg.UpdateFoo("new foo")
		log.Println("persist aggregate")
		err = anotherAgg.Store(eventStore)
		if assert.Nil(T, err) {
			log.Println("err was nil on store of aggregate")
		}
		assert.Equal(T, 0, len(anotherAgg.Events))
	})

	When(`^we retrieve the events for the aggregate$`, func() {
		var err error
		events, err = eventStore.RetrieveEvents(anotherAgg.ID)
		if err != nil {
			log.Infof("error retrieving events for the aggregate: %s", err.Error())
		}
		assert.Nil(T, err)
		assert.Equal(T, 2, len(events))
	})

	Then(`^all the events for the aggregate are returned in order$`, func() {
		assert.Equal(T, TestAggCreatedTypeCode, events[0].TypeCode)
		assert.Equal(T, TestAggFooUpdateTypeCode, events[1].TypeCode)
	})

	Then(`^we can recrete the aggregate from the event history$`, func() {
		restored := NewTestAggFromHistory(events)
		assert.NotNil(T, restored)
		assert.Equal(T, "new foo", restored.Foo)
		assert.Equal(T, "bar2", restored.Bar)
		assert.Equal(T, "baz2", restored.Baz)
	})
}

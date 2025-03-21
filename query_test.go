package donburi_test

import (
	"testing"
	"time"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

type orderableComponentTest struct {
	time.Time
}

func (o orderableComponentTest) Order() int {
	return int(time.Since(o.Time).Milliseconds())
}

var (
	queryTagA     = donburi.NewTag()
	queryTagB     = donburi.NewTag()
	queryTagC     = donburi.NewTag()
	orderableTest = donburi.NewComponentType[orderableComponentTest]()
)

func TestQueryInQuery(t *testing.T) {
	world := donburi.NewWorld()
	world.Create(queryTagA)
	world.Create(queryTagC)
	world.Create(queryTagA, queryTagB)
	world.Create(queryTagA, queryTagB, queryTagC)

	query := donburi.NewQuery(filter.Contains(queryTagA, queryTagB))
	count := 0

	for entry := range query.Iter(world) {
		count++
		if entry.Archetype().Layout().HasComponent(queryTagA) == false {
			t.Errorf("PlayerTag should be in ent archetype")
		}
		innerQuery := donburi.NewQuery(filter.Contains(queryTagA, queryTagB, queryTagC))
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic should not happen")
			}
		}()
		for innerEntry := range innerQuery.Iter(world) {
			if innerEntry.Archetype().Layout().HasComponent(queryTagA) == false {
				t.Errorf("PlayerTag should be in ent archetype")
			}
		}
	}
}

func TestQuery(t *testing.T) {
	world := donburi.NewWorld()
	world.Create(queryTagA)
	world.Create(queryTagC)
	world.Create(queryTagA, queryTagB)

	query := donburi.NewQuery(filter.Contains(queryTagA))
	count := 0

	for entry := range query.Iter(world) {
		count++
		if entry.Archetype().Layout().HasComponent(queryTagA) == false {
			t.Errorf("PlayerTag should be in ent archetype")
		}
	}

	if count != 2 {
		t.Errorf("counter should be 2, but got %d", count)
	}
}

var _ donburi.IOrderable = orderableData{}

type orderableData struct {
	Index int
}

func (o orderableData) Order() int {
	return o.Index
}

var orderable = donburi.NewComponentType[orderableData]()

func TestOrderedQuery(t *testing.T) {
	orderedEntitiesQuery := donburi.NewOrderedQuery[orderableData](
		filter.Contains(orderable),
	)

	world := donburi.NewWorld()
	for _, i := range []int{3, 1, 2} {
		e := world.Create(orderable)
		entr := world.Entry(e)
		donburi.SetValue(entr, orderable, orderableData{i})
	}

	var i int
	for e := range orderedEntitiesQuery.IterOrdered(world, orderable) {
		o := orderable.GetValue(e)
		i += 1
		if o.Index != i {
			t.Errorf("expected %d, but got %d", i, o.Index)
		}
	}
}

func BenchmarkQuery_EachOrdered(b *testing.B) {
	world := donburi.NewWorld()
	for i := 0; i < 30000; i++ {
		e := world.Create(orderableTest)
		entr := world.Entry(e)
		donburi.SetValue(entr, orderableTest, orderableComponentTest{time.Now()})
	}

	query := donburi.NewQuery(filter.Contains(orderableTest))
	orderedQuery := donburi.NewOrderedQuery[orderableComponentTest](filter.Contains(orderableTest))
	countNormal := 0
	countOrdered := 0
	b.Run("Each", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for range query.Iter(world) {
				countNormal++
			}
		}
	})
	b.Run("EachOrdered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			orderedQuery.EachOrdered(world, orderableTest, func(entry *donburi.Entry) {
				countOrdered++
			})
		}
	})
}

func BenchmarkQuery_OnlyEachOrdered(b *testing.B) {
	world := donburi.NewWorld()
	for i := 0; i < 30000; i++ {
		e := world.Create(orderableTest)
		entr := world.Entry(e)
		donburi.SetValue(entr, orderableTest, orderableComponentTest{time.Now()})
	}

	orderedQuery := donburi.NewOrderedQuery[orderableComponentTest](filter.Contains(orderableTest))
	countOrdered := 0
	b.Run("EachOrdered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			orderedQuery.EachOrdered(world, orderableTest, func(entry *donburi.Entry) {
				countOrdered++
			})
		}
	})
}

func TestQueryMultipleComponent(t *testing.T) {
	world := donburi.NewWorld()

	world.Create(queryTagA)
	world.Create(queryTagC)
	world.Create(queryTagA, queryTagB)

	query := donburi.NewQuery(filter.Contains(queryTagA, queryTagB))
	count := query.Count(world)
	if count != 1 {
		t.Errorf("counter should be 1, but got %d", count)
	}
}

func TestComplexQuery(t *testing.T) {
	createWorldFunc := func() donburi.World {
		world := donburi.NewWorld()

		world.Create(queryTagA)
		world.Create(queryTagC)
		world.Create(queryTagA, queryTagB)

		return world
	}

	var tests = []struct {
		filter        filter.LayoutFilter
		expectedCount int
	}{
		{filter.Not(filter.Contains(queryTagA)), 1},
		{filter.And(filter.Contains(queryTagA), filter.Not(filter.Contains(queryTagB))), 1},
		{filter.Or(filter.Contains(queryTagA), filter.Contains(queryTagC)), 3},
	}

	for _, tt := range tests {
		world := createWorldFunc()
		query := donburi.NewQuery(tt.filter)
		count := query.Count(world)
		if count != tt.expectedCount {
			t.Errorf("counter should be %d, but got %d", tt.expectedCount, count)
		}
	}
}

func TestFirstEntity(t *testing.T) {
	world := donburi.NewWorld()
	world.Create(queryTagA)
	world.Create(queryTagC)
	world.Create(queryTagA, queryTagB)

	// find first entity withqueryTagC
	query := donburi.NewQuery(filter.Contains(queryTagC))
	entry, ok := query.First(world)
	if entry == nil || ok == false {
		t.Errorf("entry with queryTagC should not be nil")
	}

	entry.Remove()

	entry, ok = query.First(world)
	if entry != nil || ok {
		t.Errorf("entry with queryTagC should be nil")
	}
}

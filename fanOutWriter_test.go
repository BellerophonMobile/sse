package sse

// func TestGroupSink(t *testing.T) {
// 	var group GroupSink
// 	group.HistoryLimit = 2

// 	var client1, client2 testSink
// 	var expected *Event

// 	unsub1, err := group.Subscribe(&client1, "")
// 	assert.NoError(t, err)

// 	unsub2, err := group.Subscribe(&client2, "")
// 	assert.NoError(t, err)

// 	assert.NoError(t, group.Send(&Event{ID: "1", Type: "test1", Data: "foobar1"}))
// 	expected = &Event{ID: "1", Type: "test1", Data: "foobar1"}
// 	assert.Len(t, client1.events, 1)
// 	assert.Len(t, client2.events, 1)
// 	assert.Equal(t, expected, client1.events[0])
// 	assert.Equal(t, expected, client2.events[0])

// 	unsub2()

// 	assert.NoError(t, group.Send(&Event{ID: "2", Type: "test2", Data: "foobar2"}))
// 	expected = &Event{ID: "2", Type: "test2", Data: "foobar2"}
// 	assert.Len(t, client1.events, 2)
// 	assert.Len(t, client2.events, 1)
// 	assert.Equal(t, expected, client1.events[1])

// 	assert.NoError(t, group.SetRetryTime(10*time.Second))
// 	assert.Equal(t, 10*time.Second, client1.retryTime)
// 	assert.Equal(t, time.Duration(0), client2.retryTime)

// 	unsub1()

// 	assert.NoError(t, group.Send(&Event{ID: "3", Type: "test3", Data: "foobar3"}))
// 	assert.Len(t, client1.events, 2)
// 	assert.Len(t, client2.events, 1)

// 	var client3 testSink
// 	unsub3, err := group.Subscribe(&client3, "")
// 	assert.NoError(t, err)

// 	assert.Equal(t, []*Event{
// 		{ID: "2", Type: "test2", Data: "foobar2"},
// 		{ID: "3", Type: "test3", Data: "foobar3"},
// 	}, client3.events)
// 	assert.Len(t, client1.events, 2)
// 	assert.Len(t, client2.events, 1)

// 	unsub3()

// 	var client4 testSink
// 	unsub4, err := group.Subscribe(&client4, "2")
// 	assert.NoError(t, err)

// 	assert.NoError(t, group.SetRetryTime(20*time.Second))
// 	assert.Equal(t, 20*time.Second, client4.retryTime)

// 	assert.Equal(t, []*Event{
// 		{ID: "3", Type: "test3", Data: "foobar3"},
// 	}, client4.events)
// 	assert.Len(t, client1.events, 2)
// 	assert.Len(t, client2.events, 1)
// 	assert.Len(t, client3.events, 2)

// 	unsub4()

// 	var client5 testSink
// 	unsub5, err := group.Subscribe(&client5, "foobar")
// 	assert.NoError(t, err)

// 	assert.Equal(t, []*Event{
// 		{ID: "2", Type: "test2", Data: "foobar2"},
// 		{ID: "3", Type: "test3", Data: "foobar3"},
// 	}, client5.events)
// 	assert.Len(t, client1.events, 2)
// 	assert.Len(t, client2.events, 1)
// 	assert.Len(t, client3.events, 2)
// 	assert.Len(t, client4.events, 1)

// 	unsub5()

// 	var client6 testSink
// 	_, err = group.Subscribe(&client6, "")

// 	group.Close()

// 	assert.True(t, client6.closed)

// 	assert.Equal(t, 0, group.sinks.Len())
// 	assert.Equal(t, 0, group.history.Len())
// }

// func TestGroupSink_bad(t *testing.T) {
// 	var group GroupSink

// 	client1 := testSink{err: errors.New("test error")}
// 	unsub1, err := group.Subscribe(&client1, "")
// 	assert.NoError(t, err)

// 	assert.EqualError(t, group.Send(&Event{Data: "test"}), "test error")
// 	assert.NoError(t, group.Send(&Event{Data: "test"}), "test error")
// 	assert.Len(t, client1.events, 0)

// 	assert.NotPanics(t, func() { unsub1() })
// }

// func TestGroupSink_badio(t *testing.T) {
// 	client1 := EventSink{
// 		Writer: &badSink{err: errors.New("test write error")},
// 	}

// 	var group GroupSink
// 	group.HistoryLimit = 5

// 	unsub1, err := group.Subscribe(&client1, "")
// 	assert.NoError(t, err)

// 	assert.EqualError(t, group.SetRetryTime(5*time.Second), "test write error")
// 	unsub1()

// 	assert.NoError(t, group.Send(&Event{ID: "1", Type: "test1", Data: "foobar1"}))

// 	unsub1, err = group.Subscribe(&client1, "")
// 	assert.EqualError(t, err, "test write error")
// }

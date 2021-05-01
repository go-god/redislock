package redislock

import (
	"log"
	"sync"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func lock() {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Println("redis connection error: ", err)
		return
	}

	defer conn.Close()

	l := New(conn, "heige", "hello,world", 100)

	if ok, err := l.TryLock(); ok {
		log.Println("lock success")
		for i := 0; i < 10; i++ {
			log.Println("hello,i: ", i)
		}

		l.Unlock()
	} else {
		log.Println("lock fail")
		log.Println("err: ", err)
	}
}

func TestLockExpire(t *testing.T) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Println("redis connection error: ", err)
		return
	}

	defer conn.Close()

	// l := New(conn, "abc", 1.1, 100)
	l := New(conn, "abc", 102, 100)
	if ok, err := l.TryLock(); ok {
		log.Println("lock success")
		l.Unlock()
	} else {
		log.Println("lock fail,err: ", err)
	}
}

// TestRedisLock 测试枷锁操作
func TestRedisLock(t *testing.T) {
	lock()
}

// TestLock 并发操作的尝试枷锁
func TestLock(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			lock()

		}(&wg)
	}

	wg.Wait()
	log.Println("ok")
}

/**
=== RUN   TestLock
2021/05/01 20:13:05 lock success
2021/05/01 20:13:05 hello,i:  0
2021/05/01 20:13:05 hello,i:  1
2021/05/01 20:13:05 hello,i:  2
2021/05/01 20:13:05 hello,i:  3
2021/05/01 20:13:05 hello,i:  4
2021/05/01 20:13:05 hello,i:  5
2021/05/01 20:13:05 hello,i:  6
2021/05/01 20:13:05 hello,i:  7
2021/05/01 20:13:05 hello,i:  8
2021/05/01 20:13:05 hello,i:  9
2021/05/01 20:13:05 lock fail
2021/05/01 20:13:05 lock fail
2021/05/01 20:13:05 err:  <nil>
2021/05/01 20:13:05 lock fail
2021/05/01 20:13:05 err:  <nil>
2021/05/01 20:13:05 lock fail
2021/05/01 20:13:05 err:  <nil>
2021/05/01 20:13:05 err:  <nil>
2021/05/01 20:13:05 lock success
2021/05/01 20:13:05 lock fail
2021/05/01 20:13:05 err:  <nil>
2021/05/01 20:13:05 ok
--- PASS: TestLock (0.01s)
PASS
*/

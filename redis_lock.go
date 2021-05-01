package redislock

import (
	"github.com/gomodule/redigo/redis"
)

var (
	// DefaultExpire 加锁的key默认过期时间，单位s
	DefaultExpire = 10
)

// Lock lock data
type Lock struct {
	conn   redis.Conn // redis连接句柄，支持redis pool连接句柄
	expire int        // 设置加锁key的过期时间,单位s
	key    string     // 加锁的key

	// 加锁的token value
	// token type is string,int,int64等类型都可以
	token interface{}
}

// New 实例化redis分布式锁实例对象
func New(conn redis.Conn, key string, token interface{}, expire ...int) *Lock {
	entry := &Lock{
		key:   key,
		conn:  conn,
		token: token,
	}

	// 设置锁过期时间
	if len(expire) > 0 && expire[0] > 0 {
		entry.expire = expire[0]
	}

	if entry.expire <= 0 {
		entry.expire = DefaultExpire
	}

	return entry
}

// delScript lua脚本删除一个key保证原子性，采用lua脚本执行
// 保证原子性（redis是单线程），避免del删除了，其他client获得的lock
var delScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

// Unlock 释放锁采用redis lua脚步执行，成功返回nil
func (lock *Lock) Unlock() error {
	_, err := delScript.Do(lock.conn, lock.key, lock.token)
	return err
}

// TryLock 尝试加锁,如果加锁成功就返回true,nil
// 利用redis setEx Nx的原子性实现分布式锁
// SETEX 是一个原子（atomic）操作， 它可以在同一时间内完成设置值和设置过期时间这两个操作
// 当redis设置成功返回OK,所以这里需要判断数据是否存在以及结果是否是OK
func (lock *Lock) TryLock() (bool, error) {
	result, err := redis.String(lock.conn.Do("SET", lock.key, lock.token, "EX", lock.expire, "NX"))
	if err == redis.ErrNil {
		// The lock was not successful, it already exists.
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return result == "OK", nil
}
